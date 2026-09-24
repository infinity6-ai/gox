package bqzjob

import (
	"regexp"

	"github.com/infinity6-ai/gox/commonz/constraintz/optionalz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
)

var validator = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

type Job checked.Value[string]

func (m Job) Validate(v string) error {
	return validation.RegexMatch(validator, v, "Job")
}

func (m Job) Optional() optionalz.Optional[string] {
	return checked.Value[string](m).Optional()
}

func (m Job) Get() string {
	return checked.Value[string](m).Get()
}

func (m Job) String() string {
	return m.Get()
}

func New(val string) Job {
	ret := Job{}
	checked.MustSet(&ret, val)
	return ret
}

func (m Job) MarshalJSON() ([]byte, error) {
	return checked.Marshal[string](m)
}

func (m *Job) UnmarshalJSON(data []byte) error {
	return checked.Unmarshal(m, data)
}
