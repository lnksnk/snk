package io

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
)

func Fprint(out io.Writer, a ...any) (n int64, err error) {
	if out == nil {
		return
	}
	var wn int
	var nn int64
	var wrdfrom, _ = out.(io.ReaderFrom)
	defer func() {
		if x := recover(); x != nil {
			if err, _ = x.(error); err != nil {
				err = io.EOF
				return
			}
			err = fmt.Errorf("%v", x)
		}
	}()
	for _, d := range a {
	dostr:
		if sd, sk := d.(string); sk {
			if wrdfrom != nil {
				if nn, err = wrdfrom.ReadFrom(strings.NewReader(sd)); err != nil {
					return
				}
				n += nn
				continue
			}
			if wn, err = out.Write([]byte(sd)); err == nil {
				n += int64(wn)
				continue
			}
			return
		}
		if uint8arr, uint8arrk := d.([]uint8); uint8arrk {
			if wn, err = out.Write(uint8arr); err == nil {
				n += int64(wn)
				continue
			}
			return
		}
		if btarr, btarrk := d.([]byte); btarrk {
			if wn, err = out.Write(btarr); err == nil {
				n += int64(wn)
				continue
			}
			return
		}
		if int32arr, int32arrk := d.([]int32); int32arrk {
			d = string(int32arr)
			goto dostr
		}
		if r, rk := d.(io.Reader); rk {
			if _, wtk := r.(io.WriterTo); !wtk {
				if _, rnrk := r.(io.RuneReader); !rnrk {
					if rdrfrm, rdrfrmk := out.(io.ReaderFrom); rdrfrmk {
						var rn int64
						rn, err = rdrfrm.ReadFrom(r)
						n += rn
						if err != nil {
							if err != io.EOF {
								return
							}
						}
						continue
					}
					var p = make([]byte, 8912)
					for err == nil {
						if wn, err = r.Read(p); wn > 0 {
							if err == nil {
								wn, err = out.Write(p[:wn])
								n += int64(wn)
								if err != nil {
									return
								}
								continue
							}
							if err == io.EOF {
								wn, err = out.Write(p[:wn])
								n += int64(wn)
								break
							}
							return
						}
						if err != nil {
							if err != io.EOF {
								return
							}
							err = nil
						}
						break
					}
					continue
				}
			}
		}
		if rnr, rnrk := d.(io.RuneReader); rnrk {
			var rn, rs, rerr = rune(0), 0, error(nil)
			for rerr == nil {
				if rn, rs, rerr = rnr.ReadRune(); rs > 0 {
					if wn, err = out.Write([]byte(string(rn))); err != nil {
						return
					}
					n += int64(wn)
					continue
				}
				if rerr == nil {
					rerr = io.EOF
				}
			}
			if rerr != io.EOF {
				return n, rerr
			}
			continue
		}
		if wto, wtok := d.(io.WriterTo); wtok {
			if nn, err = wto.WriteTo(out); err != nil {
				return
			}
			n += nn
			continue
		}
		if fsf, fsfk := d.(fs.File); fsfk {
			if nn, err = Fprint(out, fsf); err != nil {
				fsf.Close()
				if err == io.EOF {
					n += nn
					err = nil
					continue
				}
			}
			fsf.Close()
			n += nn
			continue
		}
		if fopenf, fopenfk := d.(interface {
			Open() (io.ReadCloser, error)
		}); fopenfk {
			fc, fck := fopenf.(io.Closer)
			var r io.ReadCloser
			if r, err = fopenf.Open(); err != nil {
				if fck {
					fc.Close()
				}
				return
			}
			if nn, err = Fprint(out, r); err != nil {
				if err == io.EOF {
					n += nn
					r.Close()
					if fck {
						fc.Close()
					}
					err = nil
					continue
				}
			}
			r.Close()
			if fck {
				fc.Close()
			}
			n += nn
			continue
		}
		if d != nil {
			if wn, err = fmt.Fprint(out, d); err != nil {
				return
			}
			n += int64(wn)
		}
	}
	return
}

func Fprintln(out io.Writer, a ...any) (n int64, err error) {
	if n, err = Fprint(out, a...); err == nil {
		var fn int
		fn, err = fmt.Fprintln(out)
		if fn > 0 {
			n += int64(fn)
		}
	}
	return
}
