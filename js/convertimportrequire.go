package js

import (
	"fmt"
	"regexp"
	"strings"

	snkio "github.com/lnksnk/snk/io"
	"github.com/lnksnk/snk/template"
)

type ImportSection struct {
	DefaultImport string   // e.g., "React" from `import React, { useState } from 'react'`
	Namespace     string   // e.g., "styles" from `import * as styles from './styles.css'`
	NamedImports  []string // e.g., ["useState", "useEffect"]
	Source        string   // e.g., "./styles.css" or "react"
}

// ConvertToRequire translates the parsed components into CommonJS format
func (s *ImportSection) ConvertToRequire() string {
	// 1. Side-effect only: import 'styles.css' -> require('styles.css');
	if s.DefaultImport == "" && s.Namespace == "" && len(s.NamedImports) == 0 {
		return fmt.Sprintf("require('%s');", s.Source)
	}

	// 2. Destructured named imports: import { a, b } from 'c' -> const { a, b } = require('c');
	if len(s.NamedImports) > 0 && s.DefaultImport == "" && s.Namespace == "" {
		joinedNamed := strings.Join(s.NamedImports, ", ")
		return fmt.Sprintf("const { %s } = require('%s');", joinedNamed, s.Source)
	}

	// 3. Namespace/Wildcard import: import * as fs from 'fs' -> const fs = require('fs');
	if s.Namespace != "" {
		return fmt.Sprintf("const %s = require('%s');", s.Namespace, s.Source)
	}

	// 4. Default import: import axios from 'axios' -> const axios = require('axios');
	if s.DefaultImport != "" && len(s.NamedImports) == 0 {
		return fmt.Sprintf("const %s = require('%s');", s.DefaultImport, s.Source)
	}

	// 5. Hybrid (Default + Named): import React, { useState } from 'react';
	// This usually requires fallback or destructuring assignment in CommonJS
	joinedNamed := strings.Join(s.NamedImports, ", ")
	return fmt.Sprintf("const %s = require('%s');\nconst { %s } = %s;", s.DefaultImport, s.Source, joinedNamed, s.DefaultImport)
}

// Regular expression to match standard ES6 import variations
// e.g., import Default, { Named1, Named2 as Alias } from 'source'
var re = regexp.MustCompile(`^\s*import\s+(?:([\w\$]+)(?:\s*,\s*)?)?(?:\*\s+as\s+([\w\$]+))?(?:\s*,\s*)?(?:\{([^\}]+)\})?(?:\s*from\s*)?["']([^"']+)["']\s*;?\s*$`)

// Simpler fallback or specific regex for side-effect imports like: import 'style.css'
var sideEffectRe = regexp.MustCompile(`^\s*import\s+["']([^"']+)["']\s*;?\s*$`)

func ParseImport(importLine string) (*ImportSection, error) {
	// Regular expression to match standard ES6 import variations
	// e.g., import Default, { Named1, Named2 as Alias } from 'source'
	//re := regexp.MustCompile(`^\s*import\s+(?:([\w\$]+)(?:\s*,\s*)?)?(?:\*\s+as\s+([\w\$]+))?(?:\s*,\s*)?(?:\{([^\}]+)\})?(?:\s*from\s*)?["']([^"']+)["']\s*;?\s*$`)

	// Simpler fallback or specific regex for side-effect imports like: import 'style.css'
	//sideEffectRe := regexp.MustCompile(`^\s*import\s+["']([^"']+)["']\s*;?\s*$`)

	matches := re.FindStringSubmatch(importLine)
	if len(matches) == 0 {
		seMatches := sideEffectRe.FindStringSubmatch(importLine)
		if len(seMatches) > 0 {
			return &ImportSection{Source: seMatches[1]}, nil
		}
		return nil, fmt.Errorf("invalid or unsupported import statement: %s", importLine)
	}

	defaultImport := strings.TrimSpace(matches[1])
	namespace := strings.TrimSpace(matches[2])
	namedRaw := strings.TrimSpace(matches[3])
	source := strings.TrimSpace(matches[4])

	var namedImports []string
	if namedRaw != "" {
		for p := range strings.SplitSeq(namedRaw, ",") {
			namedImports = append(namedImports, strings.TrimSpace(p))
		}
	}

	return &ImportSection{
		DefaultImport: defaultImport,
		Namespace:     namespace,
		NamedImports:  namedImports,
		Source:        source,
	}, nil
}

// ConvertImportToRequire convert import statements to require in code
func ConvertImportToRequire(cdtocnvrt string, evtjssctn func(*ImportSection)) (code string) {
	var jsis *ImportSection
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
						evtjssctn(jsis)
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
