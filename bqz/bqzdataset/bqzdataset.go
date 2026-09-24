package bqzdataset

import (
	"regexp"

	"github.com/infinity6-ai/gox/commonz/constraintz/optionalz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
)

var validator = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

type Dataset checked.Value[string]

func (m Dataset) Validate(v string) error {
	return validation.RegexMatch(validator, v, "DatasetId")
}

func (m Dataset) Optional() optionalz.Optional[string] {
	return checked.Value[string](m).Optional()
}

func (m Dataset) Get() string {
	return checked.Value[string](m).Get()
}

func (m Dataset) String() string {
	return m.Get()
}

func New(val string) Dataset {
	ret := Dataset{}
	checked.MustSet(&ret, val)
	return ret
}

func (m Dataset) MarshalJSON() ([]byte, error) {
	return checked.Marshal[string](m)
}

func (m *Dataset) UnmarshalJSON(data []byte) error {
	return checked.Unmarshal(m, data)
}
