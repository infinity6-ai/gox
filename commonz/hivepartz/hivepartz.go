package hivepartz

import (
	"fmt"
	"slices"
	"strings"

	"github.com/infinity6-ai/gox/commonz/constraintz/optionalz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

type HiveParts struct {
	names  []string
	values map[string]string
}

func (h *HiveParts) Names() []string {
	return slices.Clone(h.names)
}

func (h *HiveParts) Optional(name string) optionalz.Optional[string] {
	ret, ok := h.values[name]
	return optionalz.New(ret, ok)
}

func (h *HiveParts) Get(name string) string {
	return h.Optional(name).Must()
}

func (h *HiveParts) Add(name string, value string) *HiveParts {
	checker.StrNotEmpty(name, "name")
	if h.values == nil {
		h.values = make(map[string]string)
	}
	h.names = append(h.names, name)
	h.values[name] = value
	return h
}

func (h *HiveParts) Format() string {
	var sb strings.Builder
	for i, name := range h.names {
		if i > 0 {
			sb.WriteRune('/')
		}
		sb.WriteString(name)
		sb.WriteRune('=')
		sb.WriteString(h.values[name])
	}
	return sb.String()
}

func (h *HiveParts) String() string {
	return h.Format()
}

func MustParse(p *pathz.Path) (*HiveParts, *pathz.Path) {
	hp, r, err := Parse(p)
	errorz.Check(err)
	return hp, r

}

func Parse(p *pathz.Path) (*HiveParts, *pathz.Path, error) {
	err := p.Validate(pathz.ValidateOptions{
		Absolute:   new(false),
		MaxParents: new(0),
		Wildchar:   true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("path unsupported: %w", err)
	}
	var ret HiveParts
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
