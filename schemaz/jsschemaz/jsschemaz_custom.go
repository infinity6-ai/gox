package jsschemaz

import (
	"reflect"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/schemaz/jstypez"
)

func Resolve(v any) (*jstypez.Typedef, *jstypez.Typedefs) {
	if v, ok := v.(jstypez.CustomType); ok {
		return validate(v.JSTypedef())
	}
	rv := reflect.ValueOf(v)
	if v, ok := v.(jstypez.CustomTypeSimple); ok {
		return &jstypez.Typedef{Head: &jstypez.Tuple{Typ: v.JSType(), GoTyp: rv.Type().String()}}, nil
	}
	if rv.IsValid() && rv.Kind() != reflect.Ptr {
		ptr := reflect.New(rv.Type())
		ptr.Elem().Set(rv)
		if ct, ok := ptr.Interface().(jstypez.CustomType); ok {
			return validate(ct.JSTypedef())
		}
		if ct, ok := ptr.Interface().(jstypez.CustomTypeSimple); ok {
			return &jstypez.Typedef{Head: &jstypez.Tuple{Typ: ct.JSType(), GoTyp: rv.Type().String()}}, nil
		}
	}

	return nil, nil
}

func validate(typedef *jstypez.Typedef, typedefs *jstypez.Typedefs) (*jstypez.Typedef, *jstypez.Typedefs) {
	checker.StrNotEmpty(typedef.Head.GoTyp, "GoTyp is required")
	return typedef, typedefs
}
