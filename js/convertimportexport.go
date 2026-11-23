package js

import (
	"strings"

	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/template"
)

// ConvertImportExports convert import and export statements
func ConvertImportExports(cdtocnvrt string, evtjsimpsctn func(*ImportSection), evtjsexpsctn func(*ExportsSection)) (code string) {
	var jsis *ImportSection
	var jses *ExportsSection
	var err error
	var rdr, _ = template.Readers(cdtocnvrt)
	cdtocnvrt = ""
	var sz int
	var re error
	var rdriter = func(fr func(rune) bool) {
	ragain:
		var r rune
		if r, sz, re = rdr.ReadRune(); sz > 0 && re == nil {
			if !fr(r) {
				return
			}
			goto ragain
		}
	}

	for cr := range rdriter {
		if snkio.IsTxt(cr) {
			if cdtocnvrt != "" {
				code += cdtocnvrt
				cdtocnvrt = ""
			}
			code += string(cr)
			var tr = cr
			var pr rune
			for cr = range rdriter {
				if cr == tr && pr != '\\' {
					code += string(cr)
					break
				}
				pr = cr
				code += string(cr)
			}
			continue
		}
		if cdtocnvrt += string(cr); strings.Contains(cdtocnvrt, "import") {
			code += cdtocnvrt[:strings.Index(cdtocnvrt, "import")]
			cdtocnvrt = cdtocnvrt[strings.Index(cdtocnvrt, "import"):]
			for cr = range rdriter {
				cdtocnvrt += string(cr)
				if cr == ';' {
					if jsis, err = ParseImport(cdtocnvrt); err == nil && jsis != nil {
						cdtocnvrt = jsis.ConvertToRequire()
						evtjsimpsctn(jsis)
						break
					}
				}
			}
			if cdtocnvrt != "" {
				code += cdtocnvrt
				cdtocnvrt = ""
			}
			continue
		}
		if strings.Contains(cdtocnvrt, "export") {
			code += cdtocnvrt[:strings.Index(cdtocnvrt, "export")]
			cdtocnvrt = cdtocnvrt[strings.Index(cdtocnvrt, "export"):]
			for cr = range rdriter {
				cdtocnvrt += string(cr)
				if cr == ';' {
					if jses, err = ParseExport(cdtocnvrt); err == nil && jses != nil {
						cdtocnvrt = jses.ConvertToCommonJS(cdtocnvrt)
						evtjsexpsctn(jses)
						break
					}
				}
			}
			if cdtocnvrt != "" {
				code += cdtocnvrt
				cdtocnvrt = ""
			}
		}
	}
	if cdtocnvrt != "" {
		code += cdtocnvrt
		cdtocnvrt = ""
	}
	return
}
