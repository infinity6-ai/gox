package hivepartz

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/infinity6-ai/gox/commonz/constraintz/optionalz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation"
)

var nameValidator = regexp.MustCompile(`^[a-z][a-z0-9\-]*`)
var valueValidator = regexp.MustCompile(`^[a-zA-Z0-9]+[a-z0-9\-]*`)

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

func (h *HiveParts) MustAdd(name string, value string) {
	err := h.Add(name, value)
	errorz.Check(err)
}

func (h *HiveParts) Add(name string, value string) error {
	if err := validation.RegexMatch(nameValidator, name, "name"); err != nil {
		return err
	}
	if err := validation.RegexMatch(valueValidator, value, "value"); err != nil {
		return err
	}
	if h.values == nil {
		h.values = make(map[string]string)
	}
	h.names = append(h.names, name)
	h.values[name] = value
	return nil
}

func (h *HiveParts) Format() *pathz.Path {
	return pathz.MustParse(h.FormatString())
}

func (h *HiveParts) FormatString() string {
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
	return h.FormatString()
}

func MustParse(p *pathz.Path) (*HiveParts, *pathz.Path) {
	hp, r, err := Parse(p)
	errorz.Check(err)
	return hp, r
}

func ParseString(s string) (*HiveParts, *pathz.Path, error) {
	p, err := pathz.Parse(s)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing path: %w", err)
	}
	return Parse(p)
}

func (h *HiveParts) Parse(p *pathz.Path) (*pathz.Path, error) {
	err := p.Validate(pathz.ValidateOptions{
		Absolute:   new(false),
		MaxParents: new(0),
		Wildchar:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("path unsupported: %w", err)
	}
	if p.PartsLen() == 0 {
		return p, nil
	}

	parts := p.Parts()

	h.names = make([]string, 0, len(parts))
	h.values = make(map[string]string, len(parts))

	parseEndIndex := 0

	for i, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 3)
			if len(kv) > 2 {
				return nil, fmt.Errorf("too many parts in hive partition: %s", part)
			}
			if len(kv) < 2 {
				break
			}
			h.Add(kv[0], kv[1])
			parseEndIndex = i + 1
		} else {
			break
		}
	}

	p = pathz.New(0, parts[parseEndIndex:], p.HasEndingSlash())
	return p, nil
}

func Parse(p *pathz.Path) (*HiveParts, *pathz.Path, error) {
	hp := &HiveParts{}
	p, err := hp.Parse(p)
	return hp, p, err
}

func New() *HiveParts {
	return &HiveParts{}
}

func From(args ...string) *HiveParts {
	if len(args)%2 != 0 {
		panic("invalid number of arguments")
	}
	hp := New()
	for i := 0; i < len(args); i += 2 {
		hp.MustAdd(args[i], args[i+1])
	}
	return hp
}
