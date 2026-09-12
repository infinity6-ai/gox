package fsz

import (
	"regexp"

	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
)

var bucketNameValidator = regexp.MustCompile(`^[a-z][a-z0-9\-]{1,}[a-z0-9]$`)

type Bucket checked.Value[string]

func (d Bucket) Get() string {
	return checked.Value[string](d).Get()
}

func (d Bucket) String() string {
	return d.Get()
}

func (d Bucket) MarshalJSON() ([]byte, error) {
	return checked.Marshal[string](d)
}

func (d Bucket) UnmarshalJSON(data []byte) error {
	return checked.Unmarshal(&d, data)
}

func (d Bucket) Validate(v string) error {
	if err := validation.StrNotEmpty(v, "cannot be empty"); err != nil {
		return err
	}
	if err := validation.RegexMatch(bucketNameValidator, v, "invalid name"); err != nil {
		return err
	}
	return nil
}

func Id(datasetId string) Bucket {
	var ret Bucket
	checked.Set(&ret, datasetId)
	return ret
}
