package rulez

import (
	"encoding/json"
	"regexp"
	"strings"
)

type Operation string

const (
	MatchOperation    Operation = "match"
	MismatchOperation Operation = "mismatch"
	ReplaceOperation  Operation = "replace"
	SplitOperation    Operation = "split"
	DeleteOperation   Operation = "delete"
)

type Rule struct {
	Name      string         `json:"name"`
	Regexp    *regexp.Regexp `json:"-"` // Exclude from direct JSON marshaling
	Operation Operation      `json:"operation"`
	Out       string         `json:"out"`
}

// jsonRule is an auxiliary struct for JSON marshaling/unmarshaling of Rule
type jsonRule struct {
	Name          string    `json:"name"`
	RegexpPattern string    `json:"regexp_pattern"`
	Operation     Operation `json:"operation"`
	Out           string    `json:"out"`
}

// MarshalJSON implements the json.Marshaler interface for Rule.
func (r *Rule) MarshalJSON() ([]byte, error) {
	if r.Regexp == nil {
		return []byte("null"), nil
	}
	jr := jsonRule{
		Name:          r.Name,
		RegexpPattern: r.Regexp.String(),
		Operation:     r.Operation,
		Out:           r.Out,
	}
	return json.Marshal(jr)
}

// UnmarshalJSON implements the json.Unmarshaler interface for Rule.
func (r *Rule) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*r = Rule{}
		return nil
	}
	jr := jsonRule{}
	if err := json.Unmarshal(data, &jr); err != nil {
		return err
	}
	compiledRegexp, err := regexp.Compile(jr.RegexpPattern)
	if err != nil {
		return err
	}
	r.Name = jr.Name
	r.Regexp = compiledRegexp
	r.Operation = jr.Operation
	r.Out = jr.Out
	return nil
}

func (r *Rule) Apply(s string) (string, bool) {
	switch r.Operation {
	case MatchOperation:
		matched := r.Regexp.MatchString(s)
		return s, matched
	case MismatchOperation:
		return s, !r.Regexp.MatchString(s)
	case ReplaceOperation:
		return r.Regexp.ReplaceAllString(s, r.Out), true
	case SplitOperation:
		splitted := r.Regexp.Split(s, -1)
		return strings.Join(splitted, r.Out), true
	case DeleteOperation:
		return r.Regexp.ReplaceAllString(s, ""), true
	default:
		return s, false
	}

}

type Rules struct {
	Rules []Rule `json:"rules"`
}

func (e *Rules) Apply(s string) (string, int, bool) {
	c := s
	for idx, rule := range e.Rules {
		var ok bool
		c, ok = rule.Apply(c)
		if !ok {
			return s, idx, false
		}
	}
	return c, -1, true
}
