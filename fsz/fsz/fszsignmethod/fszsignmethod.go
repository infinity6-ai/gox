package fszsignmethod

import (
	"fmt"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/constraintz/optionalz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"github.com/infinity6-ai/gox/commonz/validation/checked"
	"go.code.infinity6.ai/sdkgo/genjsz/jstypez"
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

func (*SignMethod) JSType() string {
	return jstypez.TypString
}

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

func Parse(id string) (SignMethod, error) {
	var ret SignMethod
	err := checked.Set(&ret, id)
	if err != nil {
		return ret, fmt.Errorf("unsupported: %s, %w", id, err)
	}
	return ret, nil
}

func MustParseOptional(id string) SignMethod {
	var ret SignMethod
	if id == "" {
		return ret
	}
	ret = MustParse(id)
	return ret
}

func MustParse(id string) SignMethod {
	var ret SignMethod
	checked.MustSet(&ret, id)
	return ret
}
