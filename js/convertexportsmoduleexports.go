package js

import (
	"fmt"
	"regexp"
	"strings"

	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/template"
)

type ExportsSection struct {
	Type          string   // "default", "named", "inline", "reexport"
	InlineKeyword string   // "const", "let", "var", "function", "class" (if inline)
	Identifier    string   // Main name for default or inline declarations
	NamedExports  []string // List of exported variables/aliases e.g., ["foo", "bar as baz"]
	Source        string   // Source module path if re-exporting (e.g., "./mod")
}

// ConvertToCommonJS translates the parsed components into module.exports syntax
func (s *ExportsSection) ConvertToCommonJS(originalLine string) string {
	switch s.Type {
	case "default":
		// export default myFunc; -> module.exports = myFunc;
		return fmt.Sprintf("module.exports = %s;", s.Identifier)

	case "inline":
		// export const foo = 1; -> const foo = 1; exports.foo = foo;
		// (Or strips "export " and appends a binding statement to preserve valid assignments)
		stripped := strings.TrimSpace(strings.Replace(originalLine, "export ", "", 1))
		return fmt.Sprintf("%s\nexports.%s = %s;", stripped, s.Identifier, s.Identifier)

	case "named":
		// export { foo, bar }; -> exports.foo = foo; exports.bar = bar;
		var results []string
		for _, name := range s.NamedExports {
			local, exported := parseAlias(name)
			results = append(results, fmt.Sprintf("exports.%s = %s;", exported, local))
		}
		return strings.Join(results, "\n")

	case "reexport":
		// export { foo } from './mod'; -> exports.foo = require('./mod').foo;
		var results []string
		for _, name := range s.NamedExports {
			local, exported := parseAlias(name)
			results = append(results, fmt.Sprintf("exports.%s = require('%s').%s;", exported, s.Source, local))
		}
		return strings.Join(results, "\n")

	default:
		return originalLine
	}
}

// parseAlias safely manages "foo as bar" mappings
func parseAlias(name string) (string, string) {
	if strings.Contains(name, " as ") {
		parts := strings.Split(name, " as ")
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return name, name
}

var defaultRe = regexp.MustCompile(`^\s*export\s+default\s+([^;\s]+)\s*;?\s*$`)
var inlineRe = regexp.MustCompile(`^\s*export\s+(const|let|var|function|class)\s+([\w\$]+)`)
var namedRe = regexp.MustCompile(`^\s*export\s*\{([^\}]+)\}(?:\s*from\s*["']([^"']+)["'])?\s*;?\s*$`)

func ParseExport(exportLine string) (*ExportsSection, error) {

	// 1. Check Default Export
	if matches := defaultRe.FindStringSubmatch(exportLine); len(matches) > 0 {
		return &ExportsSection{
			Type:       "default",
			Identifier: strings.TrimSpace(matches[1]),
		}, nil
	}

	// 2. Check Inline Variable/Function Export
	if matches := inlineRe.FindStringSubmatch(exportLine); len(matches) > 0 {
		return &ExportsSection{
			Type:          "inline",
			InlineKeyword: matches[1],
			Identifier:    matches[2],
		}, nil
	}

	// 3. Check Named Batch or Re-export
	if matches := namedRe.FindStringSubmatch(exportLine); len(matches) > 0 {
		namedRaw := matches[1]
		source := strings.TrimSpace(matches[2])

		var namedExports []string
		parts := strings.Split(namedRaw, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				namedExports = append(namedExports, trimmed)
			}
		}

		exportType := "named"
		if source != "" {
			exportType = "reexport"
		}

		return &ExportsSection{
			Type:         exportType,
			NamedExports: namedExports,
			Source:       source,
		}, nil
	}

	return nil, fmt.Errorf("invalid or unsupported export statement")
}

// ConvertToExports convert code to CommonJS emport commands
func ConvertToExports(cdtocnvrt string, evtjsexpsctn func(*ExportsSection)) (code string) {
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
		if cdtocnvrt += string(cr); strings.Contains(cdtocnvrt, "export") {
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
