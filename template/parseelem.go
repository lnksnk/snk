package template

import (
	"io"
	"iter"

	snkio "github.com/lnksnk/snk/io"
)

type ElemType int

const (
	ElemStart ElemType = iota
	ElemSingle
	ElemEnd
)

func FparseElem(out io.Writer, a ...any) (err error) {
	var rr RuneReaders
	if rr, err = Readers(a...); err == nil {
		defer rr.Reset()
		err = ParseMarkup(rr, func(cntnt snkio.BufferWriter) (cnterr error) {
			_, cnterr = cntnt.WriteTo(out)
			return
		}, func(flushcontent func() error, elmtype ElemType, ename string, eargs map[string][]rune) (elmerr error) {
			return flushcontent()
		})

	}
	return
}

func ParseMarkup(rnsr RuneReaders, foundContent func(cntnt snkio.BufferWriter) (cnterr error), foundElem func(flushcontent func() error, elmtype ElemType, ename string, eargs map[string][]rune) (elmerr error)) (prerr error) {
	var pr rune
	var cntnt snkio.BufferWriter
	var enm []rune
	var eargs map[string][]rune

	var etp ElemType
	var trns []rune
	var nxtpr = func() (r rune, rerr error) {
		if r, _, rerr = rnsr.ReadRune(); rerr == nil && len(trns) > 0 {
			trns = append(trns, r)
		}
		return
	}

	var nextiter iter.Seq[rune] = func(yield func(rune) bool) {
		var nr rune
	rtryitr:
		if nr, prerr = nxtpr(); prerr != nil {
			return
		}
		if yield(nr) {
			goto rtryitr
		}
	}

	if nextiter != nil {

	}

	var cprtrns = func(r ...rune) {
		if cntnt != nil {
			cntnt.WriteRunes(r...)
			return
		}
		cntnt, _ = snkio.WriterBuffer(r)
	}

	var flshcntnt = func() (fsherr error) {
		if len(trns) > 0 {
			cprtrns(trns...)
			trns = nil
		}
		if cntnt != nil && !cntnt.Empty() {
			fsherr = foundContent(cntnt)
			cntnt.Reset()
			return
		}
		return
	}

	var cprtarg = func(arnme string, argv ...rune) {
		argv = append([]rune{}, argv...)
		if eargs == nil {

			eargs = map[string][]rune{arnme: argv}
			return
		}
		eargs[arnme] = argv
	}
	defer func() {
		if cntnt != nil && !cntnt.Empty() {
			cntnt.Close()
		}
		trns = nil
		eargs = nil
		enm = nil
	}()
	for {
		if pr, prerr = nxtpr(); prerr == nil {
			if pr == '<' {
				if prerr = flshcntnt(); prerr == nil {
					trns = append(trns, pr)
				nxtnm0:
					if pr, prerr = nxtpr(); prerr == nil {
						if IsName(pr) {
							enm = append(enm, pr)
							goto nxtnm0
						}
						enml := len(enm)
						if enml > 0 {
							if snkio.IsSpace(pr) {
								rnsr.Print(string(pr))
								trns = trns[:len(trns)-1]
								goto cptrargs
							}
							if pr == '#' {
								if pr, prerr = nxtpr(); prerr != nil {
									return
								}
								if pr != '>' {
									if prerr = flshcntnt(); prerr == nil {
										etp = ElemStart
										enm = nil
										eargs = nil
										trns = nil
										continue
									}
									return
								}
								enm = append(enm, '#')
								enml++
							}
						}
						if pr == '>' {
							if enml > 0 {
								goto finelm
							}
							if prerr = flshcntnt(); prerr == nil {
								etp = ElemStart
								enm = nil
								eargs = nil
								trns = nil
								continue
							}
							return
						}
						if pr == '/' {
							if trnsl := len(trns); trnsl > 1 && trns[trnsl-2] == pr {
								cprtrns(trns...)
								enm = nil
								trns = nil
								continue
							}
							if enml == 0 {
								etp = ElemEnd
							nxtnm1:
								if pr, prerr = nxtpr(); prerr == nil {
									if IsName(pr) {
										enm = append(enm, pr)
										goto nxtnm1
									}
									if pr == '>' {
										if len(enm) == 0 {
											cprtrns(trns...)
											enm = nil
											trns = nil
											continue
										}
										goto finelm
									}
									if snkio.IsSpace(pr) {
										trns = trns[:len(trns)-1]
										goto cptrargs
									}
								}
								cprtrns(trns...)
								enm = nil
								trns = nil
								continue
							}
							etp = ElemSingle
							if pr, prerr = nxtpr(); prerr == nil {
								if pr == '>' {
									goto finelm
								}
								cprtrns(trns...)
								enm = nil
								trns = nil
								continue
							}
							return
						}
						cprtrns(trns...)
						enm = nil
						trns = nil
						continue
					finelm:
						if prerr = foundElem(flshcntnt, etp, string(enm), eargs); prerr == nil {
							etp = ElemStart
							enm = nil
							eargs = nil
							trns = nil
							continue
						}
						return
					cptrargs:
						var argnme []rune
						if pr, prerr = nxtpr(); prerr == nil {
							if pr == '/' {
								etp = ElemSingle
								if pr, prerr = nxtpr(); prerr == nil {
									if pr == '>' {
										goto finelm
									}
									goto finelm
								}
								return
							}
							if pr == '>' {
								goto finelm
							}
							if snkio.IsSpace(pr) {
								goto cptrargs
							}
							if IsName(pr) {
								argnme = append(argnme, pr)
							chkargnm0:
								if pr, prerr = nxtpr(); prerr == nil {
									if IsName(pr) {
										argnme = append(argnme, pr)
										goto chkargnm0
									}
									if snkio.IsSpace(pr) {
										cprtarg(string(argnme))
										goto cptrargs
									}
									if pr == '=' {
										goto chkeq
									}
									cprtrns(trns...)
									enm = nil
									trns = nil
									continue
								}
								return
							}
							if pr == '=' {
								goto chkeq
							}
							cprtrns(trns...)
							enm = nil
							trns = nil
							continue
						chkeq:
							if pr, prerr = nxtpr(); prerr == nil {
								if snkio.IsSpace(pr) {
									goto chkeq
								}
								goto cptrval
							}
							continue
						cptrval:
							var argval []rune
							if IsName(pr) {
								argval = append(argval, pr)
								for pr = range nextiter {
									if !IsName(pr) {
										cprtarg(string(argnme), argval...)
										argval = nil
										argnme = nil
										rnsr.Print(string(pr))
										trns = trns[:len(trns)-1]
										goto cptrargs
									}
									argval = append(argval, pr)
								}
								return
							}
							if snkio.IsTxt(pr) {
								var tr = pr
								for pr = range nextiter {
									if pr == tr {
										cprtarg(string(argnme), argval...)
										argval = nil
										argnme = nil
										break
									}
									argval = append(argval, pr)
								}
								if prerr == nil {
									goto cptrargs
								}
								return
							nxttxt:
								if pr, prerr = nxtpr(); prerr == nil {
									if pr == tr {
										cprtarg(string(argnme), argval...)
										argval = nil
										argnme = nil
										goto cptrargs
									}
									argval = append(argval, pr)
									goto nxttxt
								}
								return
							}
							if pr == '[' {
								// [$---$]
								if pr, prerr = nxtpr(); prerr == nil {
									if pr == '$' {
										for {
											if pr, prerr = nxtpr(); prerr == nil {
												if pr == '$' {
													if pr, prerr = nxtpr(); prerr == nil {
														if pr == ']' {
															cprtarg(string(argnme), argval...)
															argval = nil
															argnme = nil
															goto cptrargs
														}
														argval = append(argval, '$', pr)
														continue
													}
													return
												}
												argval = append(argval, pr)
												continue
											}
											return
										}
									}
									cprtrns(trns...)
									enm = nil
									trns = nil
									continue
								}
								return
							}
							//todo [$..$]
							cprtrns(trns...)
							enm = nil
							trns = nil
							continue
						}
					}
				}
				break
			}
			cprtrns(pr)
			continue
		}
		if prerr == io.EOF {
			if cntnt != nil && !cntnt.Empty() {
				if prerr = flshcntnt(); prerr == nil {
					continue
				}
				return
			}
			prerr = nil
			return
		}
		return
	}
	return
}
