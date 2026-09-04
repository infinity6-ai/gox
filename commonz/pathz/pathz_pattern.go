package pathz

import (
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// PathPattern represents a parsed path pattern.
// It is internal and used to optimize pattern matching and formatting.
type PathPattern struct {
	originalPattern *Path
	segments        []patternSegment
	paramNames      map[string]struct{} // Set of parameter names in the pattern
}

// patternSegment represents a part of the path pattern, either a literal string or a parameter.
type patternSegment struct {
	isParam bool
	value   string // Literal string or parameter name
}

func (p *Path) ParsePattern(pattern *Path) (map[string]string, error) {
	if p.parents != pattern.parents {
		return nil, fmt.Errorf("path parents mismatch: path parents %d, pattern parents %d", p.parents, pattern.parents)
	}
	if p.hasEndingSlash != pattern.hasEndingSlash {
		return nil, fmt.Errorf("path ending slash mismatch: path has ending slash %t, pattern has ending slash %t", p.hasEndingSlash, pattern.hasEndingSlash)
	}
	if len(p.parts) != len(pattern.parts) {
		return nil, fmt.Errorf("path length mismatch: path has %d parts, pattern has %d parts", len(p.parts), len(pattern.parts))
	}

	params := make(map[string]string)

	for i := range p.parts {
		patternPart := pattern.parts[i]
		pathPart := p.parts[i]

		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			paramName := patternPart[1 : len(patternPart)-1]
			if paramName == "" {
				return nil, fmt.Errorf("empty parameter name in pattern part: '%s'", patternPart)
			}
			if strings.ContainsRune(paramName, '/') || strings.ContainsRune(paramName, '{') || strings.ContainsRune(paramName, '}') {
				return nil, fmt.Errorf("illegal character in parameter name '%s'", paramName)
			}
			
			if _, exists := params[paramName]; exists {
				return nil, fmt.Errorf("duplicate parameter name found: '%s'", paramName)
			}
			params[paramName] = pathPart
		} else {
			if patternPart != pathPart {
				return nil, fmt.Errorf("literal part mismatch at index %d: expected '%s', got '%s'", i, patternPart, pathPart)
			}
		}
	}

	return params, nil
}

func (p *Path) MustParsePattern(pattern *Path) map[string]string {
	params, err := p.ParsePattern(pattern)
	errorz.Check(err)
	return params
}

func (p *Path) FormatPattern(pattern *Path, params map[string]string) (*Path, error) {
	newParts := make([]string, 0, len(pattern.parts))
	usedParams := make(map[string]struct{})

	for i, patternPart := range pattern.parts {
		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			paramName := patternPart[1 : len(patternPart)-1]
			if paramName == "" {
				return nil, fmt.Errorf("empty parameter name in pattern part: '%s'", patternPart)
			}
			if strings.ContainsRune(paramName, '/') || strings.ContainsRune(paramName, '{') || strings.ContainsRune(paramName, '}') {
				return nil, fmt.Errorf("illegal character in parameter name '%s'", paramName)
			}

			value, ok := params[paramName]
			if !ok {
				return nil, fmt.Errorf("missing parameter value for '%s' in pattern part index %d ('%s')", paramName, i, patternPart)
			}
			newParts = append(newParts, value)
			usedParams[paramName] = struct{}{}
		} else {
			newParts = append(newParts, patternPart)
		}
	}

	for paramName := range params {
		if _, ok := usedParams[paramName]; !ok {
			return nil, fmt.Errorf("extra parameter provided: '%s' not found in pattern", paramName)
		}
	}

	newPath := New(pattern.parents, newParts, pattern.hasEndingSlash)
	return newPath, nil
}

func (p *Path) MustFormatPattern(pattern *Path, params map[string]string) *Path {
	newPath, err := p.FormatPattern(pattern, params)
	errorz.Check(err)
	return newPath
}
