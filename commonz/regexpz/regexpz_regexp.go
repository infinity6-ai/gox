package regexpz

import (
	"encoding/json"
	"regexp"
)

// Regexp is a wrapper around regexp.Regexp that supports JSON marshalling and unmarshalling.
type Regexp struct {
	*regexp.Regexp
}

// NewRegexp creates a new Regexp from a pattern string.
func NewRegexp(pattern string) (*Regexp, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &Regexp{re}, nil
}

// MustCompile is like regexp.MustCompile but for regexpz.Regexp.
func MustCompile(pattern string) *Regexp {
	re, err := NewRegexp(pattern)
	if err != nil {
		panic(err)
	}
	return re
}

// MarshalJSON implements the json.Marshaler interface.
func (r Regexp) MarshalJSON() ([]byte, error) {
	if r.Regexp == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(r.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Regexp) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		r.Regexp = nil
		return nil
	}
	re, err := regexp.Compile(s)
	if err != nil {
		return err
	}
	r.Regexp = re
	return nil
}
