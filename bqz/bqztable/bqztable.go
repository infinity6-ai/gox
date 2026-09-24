package bqztable

import (
	"regexp"

	"github.com/infinity6-ai/gox/commonz/constraintz/optionalz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
)

var validator = regexp.MustCompile(`^[a-z][a-z0-9_\-]*[a-z0-9]$`)

type Table checked.Value[string]

func (m Table) Validate(v string) error {
	return validation.RegexMatch(validator, v, "Table")
}

func (m Table) Optional() optionalz.Optional[string] {
	return checked.Value[string](m).Optional()
}

func (m Table) Get() string {
	return checked.Value[string](m).Get()
}

func (m Table) String() string {
	return m.Get()
}

func New(val string) Table {
	ret := Table{}
	checked.MustSet(&ret, val)
	return ret
}

func (m Table) MarshalJSON() ([]byte, error) {
	return checked.Marshal[string](m)
}

func (m *Table) UnmarshalJSON(data []byte) error {
	return checked.Unmarshal(m, data)
}
