package template

import (
	"fmt"
	"regexp"
	"strings"
)

func ConvertImportToRequire(jsCode string) string {
	lines := strings.Split(jsCode, "\n")
	var result []string

	// Regex for imports with a 'from' clause
	importRegex := regexp.MustCompile(`^(\s*)import\s+(.+?)\s+from\s+['"]([^'"]+)['"]\s*;?`)
	// Regex for side-effect imports
	sideEffectRegex := regexp.MustCompile(`^(\s*)import\s+['"]([^'"]+)['"]\s*;?`)

	for _, line := range lines {
		if match := importRegex.FindStringSubmatch(line); match != nil {
			indent := match[1]
			importedItem := strings.TrimSpace(match[2])
			modulePath := match[3]

			var converted string

			// Handle Hybrid Import: e.g., "React, { useState, useEffect }"
			if strings.Contains(importedItem, ",") && strings.Contains(importedItem, "{") {
				parts := strings.SplitN(importedItem, ",", 2)
				defaultVar := strings.TrimSpace(parts[0])
				namedVars := strings.TrimSpace(parts[1])

				// Generates: const React = require('react'), { useState, useEffect } = React;
				converted = fmt.Sprintf("%sconst %s = require('%s'), %s = %s;", indent, defaultVar, modulePath, namedVars, defaultVar)

			} else if strings.HasPrefix(importedItem, "* as ") {
				// Namespace import
				varName := strings.TrimPrefix(importedItem, "* as ")
				converted = fmt.Sprintf("%sconst %s = require('%s');", indent, varName, modulePath)

			} else if strings.HasPrefix(importedItem, "{") {
				// Named imports only
				converted = fmt.Sprintf("%sconst %s = require('%s');", indent, importedItem, modulePath)

			} else {
				// Plain default import
				converted = fmt.Sprintf("%sconst %s = require('%s');", indent, importedItem, modulePath)
			}

			result = append(result, converted)
		} else if match := sideEffectRegex.FindStringSubmatch(line); match != nil {
			indent := match[1]
			modulePath := match[2]
			converted := fmt.Sprintf("%srequire('%s');", indent, modulePath)
			result = append(result, converted)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
