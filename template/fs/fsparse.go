package fs

import (
	"io"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/template"
)

type nextelem struct {
	elmbf   snkio.BufferWriter
	elmfi   fs.FileInfo
	elmname string
	elmargs map[string]any
}

func (nxte *nextelem) Set(argnme string, argv any) {
	if elmargs := nxte.elmargs; elmargs != nil {
		elmargs[argnme] = argv
		return
	}
	nxte.elmargs = map[string]any{argnme: argv}
}

func (nxte *nextelem) WriteTo(w io.Writer) (n int64, err error) {
	if elmbf := nxte.elmbf; elmbf != nil && !elmbf.Empty() {
		return elmbf.WriteTo(w)
	}
	return
}

func (nxte *nextelem) Close() (err error) {
	elmbf := nxte.elmbf
	nxte.elmbf = nil
	if elmbf != nil && !elmbf.Empty() {
		elmbf.Close()
	}
	return
}

func FParseFS(topout io.Writer, unmatched map[string]bool, matched map[string]time.Time, topfi fs.FileInfo, fsstat fs.StatFileSystem, a ...any) (fserr error) {
	var rr template.RuneReaders
	var argsprefix string = "[#"
	var argpostfix string = "#]"
	var topargs map[string]any
	var fsopen fs.OpenFile
	var elmfi fs.FileInfo
	var elmcnt int
	var elmnxts []*nextelem
	var elmstrds []string
	var orgout = topout
	var prevename string
	ai := 0
	if al := len(a); al > 0 {
		if rr, _ = a[0].(template.RuneReaders); rr != nil {
			a = a[1:]
			al--
		}
		for ai < al {
			if argsd, argsk := a[ai].(map[string]any); argsk {
				if len(argsd) > 0 {
					if topargs == nil {
						topargs = map[string]any{}
					}
					for k, v := range argsd {
						topargs[k] = v
					}
				}
				a = append(a[:ai], a[ai+1:]...)
				al--
				continue
			}
			if fsstatd, fsstatk := a[ai].(fs.StatFileSystem); fsstatk {
				if fsstat == nil {
					fsstat = fsstatd
				}
				if fsopen == nil {
					fsopen, _ = a[ai].(fs.OpenFile)
				}
				a = append(a[:ai], a[ai+1:]...)
				al--
				continue
			}
			if fsopend, fsopenk := a[ai].(fs.OpenFile); fsopenk {
				if fsopen == nil {
					fsopen = fsopend
				}
				if fsstat == nil {
					fsstat, _ = a[ai].(fs.StatFileSystem)
				}
				a = append(a[:ai], a[ai+1:]...)
				al--
				continue
			}
			ai++
		}
	}
	if fsopen == nil && fsstat != nil {
		fsopen, _ = fsstat.(fs.OpenFileSystem)
	}

	if len(a) == 0 {
		var f, _ = fsopen.Open(topfi.Path())
		if f != nil {
			a = append(a, f)
		}
	}
	if rr != nil {
		if fserr = rr.Print(a...); fserr != nil {
			return
		}
	} else {
		if rr, fserr = template.Readers(a...); fserr != nil {
			return
		}
		defer rr.Reset()
	}
	if topargs == nil {
		topargs = map[string]any{}
	}
	topargs["ep-path"] = topfi.Path()
	topargs["ep-root"] = topfi.Root()
	topargs["ep-base"] = topfi.Base()
	topargs["e-root"] = strings.ReplaceAll(topfi.Root(), "/", ":")
	topargs["e-base"] = strings.ReplaceAll(topfi.Base(), "/", ":")
	var matchedinline map[string]bool
	if len(topargs) > 0 {
		rr = &argsReaders{args: topargs, RuneReaders: rr, prerns: []rune(argsprefix), postrns: []rune(argpostfix)}
	}
	if fserr = template.ParseMarkup(rr, func(cntnt snkio.BufferWriter) (cnterr error) {
		cntnt.WriteTo(topout)
		return
	}, func(flushcontent func() error, elmtype template.ElemType, ename string, eargs map[string][]rune) (elmerr error) {
		var elemname, elempath, elemext = elempth(topfi, ename, elmtype, eargs, topargs, elmnxts)
		if elempath == topfi.Path() {
			return flushcontent()
		}
		if elmtype == template.ElemEnd {
			if elmcnt > 0 && elmstrds[elmcnt-1] == elemname {
				elmstrds = elmstrds[:elmcnt-1]
				var emnxt = elmnxts[elmcnt-1]
				defer emnxt.Close()
				elmnxts = elmnxts[:elmcnt-1]
				if matchedinline[elempath] {
					delete(matchedinline, elempath)
					if elmcnt--; elmcnt > 0 {
						topout = elmnxts[elmcnt-1].elmbf
						if !emnxt.elmbf.Empty() {
							elmbfr, _ := emnxt.elmbf.Reader()
							elmnxts[elmcnt-1].Set(ename, elmbfr)
						}
						return
					}
					topout = orgout
					if !emnxt.elmbf.Empty() {
						if !emnxt.elmbf.Empty() {
							elmbfr, _ := emnxt.elmbf.Reader()
							if topargs == nil {
								topargs = map[string]any{ename: elmbfr}
								return
							}
							topargs[ename] = elmbfr
							return
						}
					}
					return
				}
				if !emnxt.elmbf.Empty() {
					if emnxt.elmargs == nil {
						emnxt.elmargs = map[string]any{"cntnt": emnxt.elmbf.Clone()}
						emnxt.elmbf.Reset()
					} else {
						emnxt.elmargs["cntnt"] = emnxt.elmbf.Clone()
						emnxt.elmbf.Reset()
					}
				}
				emnxt.elmargs["e-root"] = strings.ReplaceAll(emnxt.elmfi.Root(), "/", ":")
				emnxt.elmargs["e-base"] = strings.ReplaceAll(emnxt.elmfi.Base(), "/", ":")
				emnxt.elmargs["ep-path"] = emnxt.elmfi.Path()
				emnxt.elmargs["ep-root"] = emnxt.elmfi.Root()
				emnxt.elmargs["ep-base"] = emnxt.elmfi.Base()
				if elmcnt--; elmcnt > 0 {
					topout = elmnxts[elmcnt-1].elmbf
					if len(emnxt.elmargs) >= 0 {
						if elmerr = FParseFS(topout, unmatched, matched, emnxt.elmfi, fsstat, fsopen, emnxt.elmargs); elmerr == nil {
							emnxt.elmargs = nil
							_, elmerr = emnxt.elmbf.WriteTo(topout)
						}
						emnxt.elmargs = nil
						return
					}
					if elmerr = FParseFS(topout, unmatched, matched, emnxt.elmfi, fsstat, fsopen); elmerr == nil {
						_, elmerr = emnxt.elmbf.WriteTo(topout)
					}
					return
				}
				topout = orgout
				if len(emnxt.elmargs) > 0 {
					if elmerr = FParseFS(topout, unmatched, matched, emnxt.elmfi, fsstat, fsopen, emnxt.elmargs); elmerr == nil {
						emnxt.elmargs = nil
						_, elmerr = emnxt.elmbf.WriteTo(topout)
					}
					emnxt.elmargs = nil
					return
				}
				if elmerr = FParseFS(topout, unmatched, matched, emnxt.elmfi, fsstat, fsopen); elmerr == nil {
					_, elmerr = emnxt.elmbf.WriteTo(topout)
				}
				return
			}
			return flushcontent()
		}

		if elmfi, elmerr = func() (fs.FileInfo, error) {
			if elml := len(elemname); elemname[elml-1] == '#' {
				elemname = elemname[:elml-1]
				elmfi = fs.GenFileInfo(elempath, topfi.Root(), 0, false, time.Now())
				if matchedinline == nil {
					matchedinline = map[string]bool{strings.Replace(elmfi.Path(), "#.", ".", 1): true}
					return elmfi, nil
				}
				matchedinline[strings.Replace(elmfi.Path(), "#.", ".", 1)] = true
				return elmfi, nil
			}
			return fsstat.Stat(elempath)
		}(); elmerr == nil && elmfi != nil {
			if matched != nil && !strings.Contains(elempath, "#.") {
				matched[elempath] = elmfi.ModTime()
			}
			if elmtype == template.ElemSingle && !strings.Contains(elempath, "#.") {
				if elemname != "" && elemext != "" {
					var bfw, _ = snkio.WriterBuffer()
					defer bfw.Close()
					var elmargs = map[string]any{}
					if len(eargs) > 0 {
						elmargs = map[string]any{}
						for k, v := range eargs {
							elmargs[k] = string(v)
						}
					}
					elmargs["e-root"] = strings.ReplaceAll(elmfi.Root(), "/", ":")
					elmargs["e-base"] = strings.ReplaceAll(elmfi.Base(), "/", ":")
					elmargs["ep-path"] = elmfi.Path()
					elmargs["ep-root"] = elmfi.Root()
					elmargs["ep-base"] = elmfi.Base()
					if elmerr = FParseFS(bfw, unmatched, matched, elmfi, fsstat, fsopen, elmargs); elmerr == nil {
						elmargs = nil
						_, elmerr = bfw.WriteTo(topout)
					}
					elmargs = nil
				}
				return
			}
			if elmtype == template.ElemStart {
				var elmbf, _ = snkio.WriterBuffer()
				topout = elmbf
				elmcnt++
				elmstrds = append(elmstrds, elemname)
				var elmargs map[string]any
				if len(eargs) > 0 {
					elmargs = map[string]any{}
					for k, v := range eargs {
						elmargs[k] = string(v)
					}
				}
				elmnxts = append(elmnxts, &nextelem{elmbf: elmbf, elmname: elemname, elmfi: elmfi, elmargs: elmargs})
				return
			}
		}
		if unmatched != nil {
			unmatched[elempath] = true
		}
		if prevename != ename {
			if prevename == "head" && filepath.Ext(topfi.Path()) == ".html" {
				//generate a default base tag for html content with a head tag
				if ename != "base" {
					var basehref, _ = topargs["base-href"].(string)
					if basehref == "" {
						basehref, _ = topargs["ep-root"].(string)
					}
					if _, fserr = snkio.Fprint(topout, `<base href="`, basehref, `"/>`); fserr != nil {

					}
				}
			}
			prevename = ename
		}

		return flushcontent()
	}); fserr != nil {

	}
	return
}

func elempth(topfi fs.FileInfo, elmname string, elmtype template.ElemType, eargs map[string][]rune, topargs map[string]any, elmnxts []*nextelem) (elemname, elempath, elemext string) {
	var sip = 0
	var orgnme = elmname

retrysp:
	if sip = strings.Index(elmname, "::"); sip > -1 {
		var prename = elmname[:sip]
		elmname = elmname[len(prename)+2:]
		if sip = strings.Index(elmname, "::"); sip > -1 {
			var esec = elmname[:sip]
			elmname = elmname[len(esec)+2:]
			if el := len(elmnxts); el > 0 {
				for k, v := range elmnxts[el-1].elmargs {
					if k == esec {
						if bfw, _ := snkio.WriterBuffer(v); bfw != nil && !bfw.Empty() {
							elemname = prename + bfw.String() + elmname
							bfw.Close()
							goto retrysp
						}
					}
				}
			}

			for k, v := range topargs {
				if k == esec {
					if bfw, _ := snkio.WriterBuffer(v); bfw != nil && !bfw.Empty() {
						elmname = prename + bfw.String() + elmname
						bfw.Close()
						goto retrysp
					}
				}
			}

			for k, v := range eargs {
				if k == esec {
					elmname = prename + string(v) + elmname
					goto retrysp
				}
			}

			elemname = orgnme
		}
	}
	elempath = strings.ReplaceAll(elmname, ":", "/")
	if elempath[0] != '/' {
		elempath = topfi.Root() + elempath
	}
	if elemext = filepath.Ext(elempath); elemext == "" {
		elemext = filepath.Ext(topfi.Path())
		li := len(elempath) - 1
		elemname = strings.ReplaceAll(elempath[:li+1], "/", ":")
		if elempath[li] == '/' {
			elempath += "index" + elemext
		} else {
			elempath += elemext
		}
		return
	} else {
		elemname = strings.ReplaceAll(elempath[:len(elempath)-len(elemext)], "/", ":")
	}
	return
}

type ArgumentsRuneReaders interface {
	template.RuneReaders
}

type argsReaders struct {
	prerns  []rune
	tmprns  []rune
	prei    int
	postrns []rune
	posti   int
	args    map[string]any
	template.RuneReaders
}

func (ar *argsReaders) ReadRune() (r rune, size int, err error) {
retryrd:
	if rdrs := ar.RuneReaders; rdrs != nil {
		if r, size, err = rdrs.ReadRune(); err == nil && r == ar.prerns[ar.prei] {
			ar.tmprns = append(ar.tmprns, r)
			if ar.prei++; ar.prei == len(ar.prerns) {
				var ininme []rune
			nxtnme:
				if r, size, err = rdrs.ReadRune(); err == nil && template.IsName(r) {
					ar.tmprns = append(ar.tmprns, r)
					ininme = append(ininme, r)
					goto nxtnme
				}
			nxtpst:
				if err == nil {
					if ar.postrns[ar.posti] == r {
						ar.tmprns = append(ar.tmprns, r)
						if ar.posti++; ar.posti == len(ar.postrns) {
							if av, ok := ar.args[string(ininme)]; ok {
								rdrs.Print(av)
								ar.tmprns = nil
								ar.prei = 0
								ar.posti = 0
								goto retryrd
							}
							ar.prei = 0
							ar.posti = 0
							ar.tmprns = nil
							goto retryrd
						}
						if r, size, err = rdrs.ReadRune(); err != nil {
							ar.tmprns = append(ar.tmprns, r)
							return
						}
						goto nxtpst
					}
					ar.tmprns = append(ar.tmprns, r)
					r = ar.tmprns[0]
					ar.tmprns = ar.tmprns[1:]
					ar.prei = 0
					ar.posti = 0
					rdrs.Print(ar.tmprns)
					ar.tmprns = nil
					goto retryrd
				}
				return
			}
			goto retryrd
		}
		if len(ar.tmprns) == 0 {
			return
		}
		if err == nil {
			ar.prei = 0
			ar.tmprns = append(ar.tmprns, r)
			r = ar.tmprns[0]
			ar.tmprns = ar.tmprns[1:]
			rdrs.Print(ar.tmprns)
			ar.tmprns = nil
			size = utf8.RuneLen(r)
		}
		return
	}
	return
}

func (ar *argsReaders) Print(a ...any) (err error) {
	if len(a) > 0 {
		rdrs := ar.RuneReaders
		if rdrs != nil {
			err = rdrs.Print(a...)
			return
		}
		if rdrs, err = template.Readers(a...); err == nil {
			ar.RuneReaders = rdrs
		}
	}
	return
}

func (ar *argsReaders) Reset() (err error) {
	if rdrs := ar.RuneReaders; rdrs != nil {
		return rdrs.Reset()
	}
	return
}

func ArgsReaders(prephase, postphrase string, a ...any) (ar ArgumentsRuneReaders, err error) {
	var rdrs template.RuneReaders
	if rdrs, err = template.Readers(a...); err == nil {
		return &argsReaders{RuneReaders: rdrs, prerns: []rune(prephase), postrns: []rune(postphrase)}, nil
	}
	return
}
