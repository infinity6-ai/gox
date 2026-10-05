package fszsignmethod

import (
	"net/http"

	"github.com/infinity6-ai/gox/commonz/constraintz/optionalz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
)

var validMethods = []string{
	http.MethodGet,
	http.MethodPut,
	http.MethodDelete,
}

type SignMethod checked.Value[string]

var (
	SignMethodGet    = MustParse(http.MethodGet)
	SignMethodPut    = MustParse(http.MethodPut)
	SignMethodDelete = MustParse(http.MethodDelete)
)

func (d SignMethod) Optional() optionalz.Optional[string] {
	return checked.Value[string](d).Optional()
}

func (d SignMethod) Get() string {
	return checked.Value[string](d).Get()
}

func (d SignMethod) String() string {
	return d.Get()
}

func (d SignMethod) MarshalJSON() ([]byte, error) {
	return checked.Marshal[string](d)
}

func (d *SignMethod) UnmarshalJSON(data []byte) error {
	return checked.Unmarshal(d, data)
}

func (d SignMethod) Validate(v string) error {
	return validation.OneOf(validMethods, v, "unsupported method")
}

func MustParse(id string) SignMethod {
	var ret SignMethod
	checked.MustSet(&ret, id)
	return ret
}
