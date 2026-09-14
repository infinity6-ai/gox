package hivepartz

import (
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"go.code.infinity6.ai/platform/validation"
)

type HivePartsV2 struct {
	names  []string          `json:"names"`
	values map[string]string `json:"indexes"`
}

func (h *HivePartsV2) Add(name string, value string) *HivePartsV2 {
	checker.StrNotEmpty(name, "name")
	checker.StrNotEmpty(value, "value")
	h.names = append(h.names, name)
	h.values[name] = value
	return h
}

func ParseV2(p *pathz.Path) (*HivePartsV2, *pathz.Path, error) {
	err := p.Validate(pathz.ValidateOptions{
		Absolute:   new(false),
		MaxParents: new(0),
		Wildchar:   true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("path unsupported: %w", err)
	}
	var ret HivePartsV2
	if p.PartsLen() == 0 {
		return &ret, p, nil
	}

	parts := p.Parts()

	parseEndIndex := 0

	for i, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 3)
			if len(kv) > 2 {
				return nil, nil, fmt.Errorf("too many parts in hive partition: %s", part)
			}
			if len(kv) < 2 {
				break
			}
			ret.Add(kv[0], kv[1])
			parseEndIndex = i + 1
		} else {
			break
		}
	}

	p = pathz.New(0, parts[parseEndIndex:], p.HasEndingSlash())
	return &ret, p, nil
}

type HivePart struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HiveParts struct {
	Parts []*HivePart `json:"parts"`
}

func (h *HiveParts) Add(name string, value string) *HiveParts {
	part := &HivePart{Name: name, Value: value}
	h.Parts = append(h.Parts, part)
	return h
}

func (h *HiveParts) Get(idx int, name string) string {
	part := h.Parts[idx]
	validation.Equal(name, part.Name, "name")
	return part.Value
}

func (h *HiveParts) ToMap() map[string]string {
	m := make(map[string]string, len(h.Parts))
	for _, part := range h.Parts {
		m[part.Name] = part.Value
	}
	return m
}

func (h *HiveParts) Format() string {
	var parts []string
	for _, part := range h.Parts {
		parts = append(parts, part.Name+"="+part.Value)
	}
	return strings.Join(parts, "/")
}

func New() *HiveParts {
	return &HiveParts{}
}

func Format(partAndValues ...string) string {
	if len(partAndValues)%2 != 0 {
		panic(fmt.Errorf("invalid number of arguments"))
	}
	hp := New()
	for i := 0; i < len(partAndValues); i += 2 {
		hp.Add(partAndValues[i], partAndValues[i+1])
	}
	return hp.Format()
}

func ParseHive(relativePath string, max int) *HiveParts {
	ret, _ := Parse(relativePath, max)
	return ret
}

func Parse(relativePath string, max int) (*HiveParts, string) {
	if relativePath == "" {
		return &HiveParts{}, ""
	}

	if strings.HasPrefix(relativePath, "/") {
		pathz.ValidateRelative(relativePath)
	}

	parts := strings.Split(relativePath, "/")
	hiveParts := &HiveParts{}

	parseEndIndex := 0
	parsedCount := 0

	for i, part := range parts {
		if max != -1 && parsedCount >= max {
			break
		}
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
				hiveParts.Parts = append(hiveParts.Parts, &HivePart{Name: kv[0], Value: kv[1]})
				parsedCount++
				parseEndIndex = i + 1
			} else {
				break
			}
		} else {
			break
		}
	}

	// if the last part of hivepart has a trailing slash, the remaining path is empty.
	if len(parts) > 1 && parseEndIndex == (len(parts)-1) && parts[parseEndIndex] == "" {
		return hiveParts, ""
	}

	remainingPath := strings.Join(parts[parseEndIndex:], "/")
	return hiveParts, remainingPath
}
