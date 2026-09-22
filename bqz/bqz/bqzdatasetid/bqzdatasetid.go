package bqzdatasetid

import (
	"regexp"

	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
)

var validator = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

type DatasetId checked.Value[string]

func (m DatasetId) Validate(v string) error {
	return validation.RegexMatch(validator, v, "DatasetId cannot be empty")
}

func (m DatasetId) Get() string {
	return checked.Value[string](m).Get()
}

func (m DatasetId) String() string {
	return m.Get()
}

func New(val string) DatasetId {
	ret := DatasetId{}
	checked.MustSet(&ret, val)
	return ret
}

func (m DatasetId) MarshalJSON() ([]byte, error) {
	return checked.Marshal[string](m)
}

func (m *DatasetId) UnmarshalJSON(data []byte) error {
	return checked.Unmarshal(m, data)
}
