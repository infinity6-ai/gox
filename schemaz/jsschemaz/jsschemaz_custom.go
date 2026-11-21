package jsschemaz

import (
	"reflect"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/schemaz/typeastz"
)

func Resolve(v any) (*typeastz.Typedef, *typeastz.Typedefs) {
	if ct, ok := v.(typeastz.CustomType); ok {
		return validate(ct.CustomType())
	}
	if ct, ok := v.(typeastz.JSTypedef); ok {
		return validate(ct.JSTypedef())
	}

	rv := reflect.ValueOf(v)
	if ct, ok := v.(typeastz.CustomTypeSimple); ok {
		return &typeastz.Typedef{Head: &typeastz.Tuple{Typ: ct.CustomTypeSimple(), GoTyp: rv.Type().String()}}, nil
	}
	if ct, ok := v.(typeastz.JSType); ok {
		return &typeastz.Typedef{Head: &typeastz.Tuple{Typ: ct.JSType(), GoTyp: rv.Type().String()}}, nil
	}

	if rv.IsValid() && rv.Kind() != reflect.Ptr {
		ptr := reflect.New(rv.Type())
		ptr.Elem().Set(rv)
		if ct, ok := ptr.Interface().(typeastz.CustomType); ok {
			return validate(ct.CustomType())
		}
		if ct, ok := ptr.Interface().(typeastz.JSTypedef); ok {
			return validate(ct.JSTypedef())
		}
		if ct, ok := ptr.Interface().(typeastz.CustomTypeSimple); ok {
			return &typeastz.Typedef{Head: &typeastz.Tuple{Typ: ct.CustomTypeSimple(), GoTyp: rv.Type().String()}}, nil
		}
		if ct, ok := ptr.Interface().(typeastz.JSType); ok {
			return &typeastz.Typedef{Head: &typeastz.Tuple{Typ: ct.JSType(), GoTyp: rv.Type().String()}}, nil
		}
	}

	return nil, nil
}

func validate(typedef *typeastz.Typedef, typedefs *typeastz.Typedefs) (*typeastz.Typedef, *typeastz.Typedefs) {
	checker.StrNotEmpty(typedef.Head.GoTyp, "GoTyp is required")
	return typedef, typedefs
}
