package template

import (
	"fmt"
	"regexp"
	"strings"
)

type JSImportSections struct {
	DefaultImport string   // e.g., "React" from `import React, { useState } from 'react'`
	Namespace     string   // e.g., "styles" from `import * as styles from './styles.css'`
	NamedImports  []string // e.g., ["useState", "useEffect"]
	Source        string   // e.g., "./styles.css" or "react"
}

// ConvertToRequire translates the parsed components into CommonJS format
func (s *JSImportSections) ConvertToRequire() string {
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

func ParseJSImport(importLine string) (*JSImportSections, error) {
	// Regular expression to match standard ES6 import variations
	// e.g., import Default, { Named1, Named2 as Alias } from 'source'
	//re := regexp.MustCompile(`^\s*import\s+(?:([\w\$]+)(?:\s*,\s*)?)?(?:\*\s+as\s+([\w\$]+))?(?:\s*,\s*)?(?:\{([^\}]+)\})?(?:\s*from\s*)?["']([^"']+)["']\s*;?\s*$`)

	// Simpler fallback or specific regex for side-effect imports like: import 'style.css'
	//sideEffectRe := regexp.MustCompile(`^\s*import\s+["']([^"']+)["']\s*;?\s*$`)

	matches := re.FindStringSubmatch(importLine)
	if len(matches) == 0 {
		seMatches := sideEffectRe.FindStringSubmatch(importLine)
		if len(seMatches) > 0 {
			return &JSImportSections{Source: seMatches[1]}, nil
		}
		return nil, fmt.Errorf("invalid or unsupported import statement: %s", importLine)
	}

	defaultImport := strings.TrimSpace(matches[1])
	namespace := strings.TrimSpace(matches[2])
	namedRaw := strings.TrimSpace(matches[3])
	source := strings.TrimSpace(matches[4])

	var namedImports []string
	if namedRaw != "" {
		parts := strings.Split(namedRaw, ",")
		for _, p := range parts {
			namedImports = append(namedImports, strings.TrimSpace(p))
		}
	}

	return &JSImportSections{
		DefaultImport: defaultImport,
		Namespace:     namespace,
		NamedImports:  namedImports,
		Source:        source,
	}, nil
}

func ConvertImportToRequire(cdtocnvrt string, evtjssctn func(*JSImportSections)) (code string) {
	var jsis *JSImportSections
	var lstci = 0
	var err error
	for cdi := range cdtocnvrt {
		if strings.HasPrefix(cdtocnvrt[cdi:], "import") {
			if cmai := strings.Index(cdtocnvrt[cdi:], ";"); cmai > -1 {
				if jsis, err = ParseJSImport(cdtocnvrt[cdi : cdi+cmai+1]); err == nil && jsis != nil {
					code += cdtocnvrt[lstci:cdi] + jsis.ConvertToRequire()
					lstci = cdi + cmai + 1
					evtjssctn(jsis)
					continue
				}
			}
		}
	}
	if lstci < len(cdtocnvrt) {
		code += cdtocnvrt[lstci:]
	}
	return
}
