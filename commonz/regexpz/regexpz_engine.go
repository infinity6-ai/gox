package regexpz

import (
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
	Regexp    *regexp.Regexp `json:"regexp"`
	Operation Operation      `json:"operation"`
	Out       string         `json:"out"`
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

type Engine struct {
	Rules []Rule `json:"rules"`
}

func (e *Engine) Apply(s string) (string, bool) {
	c := s
	for _, rule := range e.Rules {
		var ok bool
		c, ok = rule.Apply(c)
		if !ok {
			return s, false
		}
	}
	return c, true
}
