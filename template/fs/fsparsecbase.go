package fs

import (
	"io"
	"strings"

	"github.com/lnksnk/snk/fs"
	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/template"
)

func NextRuner(rdr io.RuneReader) func() (rune, error) {
	var r, size, rerr = rune(0), 0, error(nil)
	return func() (rune, error) {
		if r, size, rerr = rdr.ReadRune(); size == 0 && rerr == nil {
			rerr = io.EOF
			return r, rerr
		}
		return r, rerr
	}
}

func RuneIter(rdr io.RuneReader) func(func(rune, error) bool) {
	var nxtr = NextRuner(rdr)
	var r rune
	var err error
	return func(f func(rune, error) bool) {
		for err == nil {
			if r, err = nxtr(); r > 0 {
				if !f(r, err) {
					return
				}
			}
		}
	}
}

func FParseCbaseFS(topout, codeout io.Writer, topfi fs.FileInfo, a ...any) (fserr error) {
	var rr template.RuneReaders
	if rr, fserr = template.Readers(a...); fserr != nil {
		return
	}
	defer rr.Reset()
	var fndcde = false
	fserr = template.ParseCbased("<%", "%>", rr, func(cntnt snkio.BufferWriter, coder rune) (cnterr error) {
		if topout == nil && !fndcde {
			fndcde = true
		}
		if cntnt.Contains("{@") && cntnt.Contains("@}") {
			if !fndcde {
				fndcde = true
			}
			rdr, _ := cntnt.Reader()
			var rniter = RuneIter(rdr)
			var cntrs []rune
			var cders []rune
			var cntpre = []rune("{@")
			var cntprei int
			var cntpost = []rune("@}")
			var cntposti int
			var fncde = false
			var wrpup bool
			var init = true

			var cptrcnt = func() {
				if len(cntrs) > 0 {
					if init {
						init = false
						if fncde {
							if wrpup {
								if len(cders) == 0 {
									if coder > 0 && strings.ContainsRune("=+([", coder) {
										_, cnterr = snkio.Fprint(codeout, cntrs, "`")
										cntrs = nil
										return
									}
									_, cnterr = snkio.Fprint(codeout, cntrs, "`);")
									cntrs = nil
									return
								}
								_, cnterr = snkio.Fprint(codeout, cntrs)
								cntrs = nil
								return
							}
							if coder > 0 && strings.ContainsRune("=+([", coder) {
								_, cnterr = snkio.Fprint(codeout, "`", cntrs)
								cntrs = nil
								return
							}
							_, cnterr = snkio.Fprint(codeout, "print(`", cntrs)
							cntrs = nil
							return
						}
						if wrpup {
							if coder > 0 && strings.ContainsRune("=+([", coder) {
								_, cnterr = snkio.Fprint(codeout, "`", cntrs, "`")
								cntrs = nil
								return
							}
							_, cnterr = snkio.Fprint(codeout, "print(`", cntrs, "`);")
							cntrs = nil
							return
						}
						if coder > 0 && strings.ContainsRune("=+([", coder) {
							_, cnterr = snkio.Fprint(codeout, "`", cntrs)
							cntrs = nil
							return
						}
						_, cnterr = snkio.Fprint(codeout, "print(`", cntrs)
						cntrs = nil
						return
					}
					if wrpup {
						if fncde {
							if len(cders) == 0 {
								if coder > 0 && strings.ContainsRune("=+([", coder) {
									_, cnterr = snkio.Fprint(codeout, cntrs, "`")
									cntrs = nil
									return
								}
								_, cnterr = snkio.Fprint(codeout, cntrs, "`);")
								cntrs = nil
								return
							}
							_, cnterr = snkio.Fprint(codeout, cntrs)
							cntrs = nil
							return
						}
						if coder > 0 && strings.ContainsRune("=+([", coder) {
							_, cnterr = snkio.Fprint(codeout, cntrs, "`")
							cntrs = nil
							return
						}
						_, cnterr = snkio.Fprint(codeout, cntrs, "`);")
						cntrs = nil
						return
					}
					_, cnterr = snkio.Fprint(codeout, cntrs)
					cntrs = nil
					return
				}
			}
			var cptrcde = func() {
				if len(cders) > 0 {
					fncde = true
					if init {
						init = false
						if wrpup {
							if coder > 0 && strings.ContainsRune("=+[(", coder) {
								_, cnterr = snkio.Fprint(codeout, "`${", cders, "}`")
								cders = nil
								return
							}
							_, cnterr = snkio.Fprint(codeout, "print(`${", cders, "}`);")
							cders = nil
							return
						}
						if coder > 0 && strings.ContainsRune("=+[(", coder) {
							_, cnterr = snkio.Fprint(codeout, "`${", cders, "}")
							cders = nil
							return
						}
						_, cnterr = snkio.Fprint(codeout, "print(`${", cders, "}")
						cders = nil
						return
					}
					if wrpup {
						if coder > 0 && strings.ContainsRune("=+[(", coder) {
							_, cnterr = snkio.Fprint(codeout, "${", cders, "}`")
							cders = nil
							return
						}
						_, cnterr = snkio.Fprint(codeout, "${", cders, "}`);")
						cders = nil
						return
					}
					_, cnterr = snkio.Fprint(codeout, "${", cders, "}")
					cders = nil
					return
				}
			}
			var nr rune
			for nr = range rniter {
				if nr == cntpre[cntprei] {
					cntprei++
					for nr = range rniter {
						if nr == cntpre[cntprei] {
							if cntprei++; cntprei == len(cntpre) {
								cntprei = 0
								cptrcde()
								for nr = range rniter {
									if nr == cntpost[cntposti] {
										cntposti++
										for nr = range rniter {
											if nr == cntpost[cntposti] {
												if cntposti++; cntposti == len(cntpost) {
													cntposti = 0
													if len(cders) > 0 {
														cptrcnt()
													}
													break
												}
												continue
											}
											if cntposti > 0 {
												cders = append(cders, cntpost[:cntposti]...)
												cntposti = 0
											}
											cders = append(cders, nr)
										}
										break
									}
									cders = append(cders, nr)
								}
								break
							}
						}
						if cntprei > 0 {
							cptrcde()
							cntrs = append(cntrs, cntpre[:cntprei]...)
							cntposti = 0
						}
						cptrcde()
						cntrs = append(cntrs, nr)
						break
					}
					continue
				}
				cptrcde()
				cntrs = append(cntrs, nr)
			}
			wrpup = true
			cptrcnt()
			cptrcde()
			return
		}
		if !fndcde {
			if topout != nil {
				_, cnterr = cntnt.WriteTo(topout)
			}
			return
		}
		_, cnterr = snkio.Fprint(codeout, "print(`", cntnt, "`);")
		return
	}, func(code snkio.BufferWriter) (cdeerr error) {
		if !fndcde {
			fndcde = true
		}
		_, cdeerr = snkio.Fprint(codeout, code)
		return
	})
	return
}
