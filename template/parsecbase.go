package template

import (
	"io"

	snkio "github.com/lnksnk/snk/io"
)

func ParseCbased(prephrase, postphrase string, rnsr RuneReaders, foundContent func(cntnt snkio.BufferWriter, coder rune) (cnterr error), foundCode func(code snkio.BufferWriter) (cdeerr error)) (prerr error) {
	if rnsr == nil {
		return
	}
	var pr rune
	var cntnt snkio.BufferWriter
	var code snkio.BufferWriter

	var prerns = []rune(prephrase)
	var prel = len(prerns)
	var prei int
	var postrns = []rune(postphrase)
	var postl = len(postrns)
	var posti int
	var trns []rune
	var coder rune
	var nxtpr = func() (r rune, rerr error) {
		if r, _, rerr = rnsr.ReadRune(); rerr == nil && len(trns) > 0 {
			trns = append(trns, r)
		}
		return
	}

	var cprtrns = func(r ...rune) {
		if cntnt != nil {
			cntnt.WriteRunes(r...)
			return
		}
		cntnt, _ = snkio.WriterBuffer()
		cntnt.WriteRunes(r...)
	}

	var cprtcderns = func(r ...rune) {
		if code != nil {
			code.WriteRunes(r...)
			return
		}
		code, _ = snkio.WriterBuffer()
		code.WriteRunes(r...)
	}

	var flshcntnt = func() (fsherr error) {
		if len(trns) > 0 {
			cprtrns(trns...)
			trns = nil
		}
		if cntnt != nil && !cntnt.Empty() {
			fsherr = foundContent(cntnt, coder)
			cntnt.Reset()
		}
		coder = 0
		return
	}

	var flscde = func() (fsherr error) {
		if code != nil && !code.Empty() {
			fsherr = foundCode(code)
			code.Reset()
			return
		}
		return
	}
	defer func() {
		if code != nil && !code.Empty() {
			code.Close()
		}
		if cntnt != nil && !cntnt.Empty() {
			cntnt.Close()
		}
		trns = nil
	}()
	for {
		if pr, prerr = nxtpr(); prerr == nil {
			if prel > 0 && postl > 0 && prerns[prei] == pr {
				if prei++; prei == prel {
				nxtpost:
					if pr, prerr = nxtpr(); prerr == nil {
						if postl > 0 && postrns[posti] == pr {
							if posti++; posti == postl {
								prei = 0
								posti = 0
								if prerr = flshcntnt(); prerr == nil {
									if prerr = flscde(); prerr == nil {
										continue
									}
								}
								return
							}
							goto nxtpost
						}
						if posti > 0 {
							cprtcderns(postrns[:posti]...)
							posti = 0
						}
						if !snkio.IsSpace(pr) {
							coder = pr
						}
						if snkio.IsTxt(pr) {
							var txtr = pr
							var ustxt = pr
							var prvr = pr
							var cdetxt []rune
							for {
								if pr, prerr = nxtpr(); prerr == nil {
									if pr == txtr {
										if prvr != '\\' {
											cprtcderns(ustxt)
											cprtcderns(cdetxt...)
											cprtcderns(ustxt)
											goto nxtpost
										}
									}
									if pr == '{' {
										if pr, prerr = nxtpr(); prerr == nil {
											if pr == '@' {
												ustxt = '`'
												cdetxt = append(cdetxt, '$', '{')
												for {
													if pr, prerr = nxtpr(); prerr == nil {
														if pr == '@' {
															if pr, prerr = nxtpr(); prerr == nil {
																if pr == '}' {
																	goto nextcdr
																}
															}
															return
														}
														cdetxt = append(cdetxt, pr)
														continue
													}
													return
												}
											}
											cdetxt = append(cdetxt, '{')
											goto nextcdr
										}
										return
									}
								nextcdr:
									if ustxt == txtr {
										cdetxt = append(cdetxt, pr)
										prvr = pr
										continue
									}
									if pr == ustxt {
										cdetxt = append(cdetxt, '\\', pr)
										prvr = pr
										continue
									}
									cdetxt = append(cdetxt, pr)
									prvr = pr
									continue
								}
								return
							}
						}
						cprtcderns(pr)
						goto nxtpost
					}
					return
				}
				continue
			}
			if prei > 0 {
				cprtrns(prerns[:prei]...)
				prei = 0
			}
			cprtrns(pr)
			continue
		}
		if prerr == io.EOF {
			if (code == nil || code.Empty()) && cntnt != nil && !cntnt.Empty() {
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
}
