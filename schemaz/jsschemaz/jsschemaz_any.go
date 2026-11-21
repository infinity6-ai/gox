package jsschemaz

import (
	"context"
	"fmt"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/jsonz/jsonzembedded"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/schemaz/jstypez"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

func (p *Parser) ParseAny(ctx context.Context, baseId string, v any) (*jstypez.Typedef, *jstypez.Typedefs) {
	if v == nil {
		return &jstypez.Typedef{Head: &jstypez.Tuple{Typ: "any", GoTyp: "any"}}, nil
	}

	val := reflect.ValueOf(v)

	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			val = reflect.Zero(val.Type().Elem())
			break
		}
		val = val.Elem()
	}

	if v, ok := jsonzembedded.Unembed(val.Interface()); ok {
		a, b := p.ParseAny(ctx, baseId, v)
		a.Head.Name = baseId
		if b == nil {
			b = &jstypez.Typedefs{}
		}
		b.Add(a)
		a = &jstypez.Typedef{Head: &jstypez.Tuple{Typ: "i6JsonEmbedded<" + a.Head.Name + ">", GoTyp: ""}}
		return a, b
	}

	if a, b := p.knownTypes(ctx, baseId, val); a != nil {
		return a, b
	}

	if a, b := p.checkSchema(ctx, baseId, val); a != nil {
		return a, b
	}

	if a, b := checkCustom(ctx, baseId, val); a != nil {
		return a, b
	}

	if !val.IsValid() {
		panic(fmt.Sprintf("invalid value or nil pointer: %#v", v))
	}

	switch t := val.Kind(); t {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return &jstypez.Typedef{Head: &jstypez.Tuple{Typ: jstypez.TypNumber, GoTyp: t.String()}}, nil

	case reflect.String:
		return &jstypez.Typedef{Head: &jstypez.Tuple{Typ: jstypez.TypString, GoTyp: t.String()}}, nil

	case reflect.Bool:
		return &jstypez.Typedef{Head: &jstypez.Tuple{Typ: jstypez.TypBoolean, GoTyp: t.String()}}, nil

	case reflect.Slice, reflect.Array:
		var allTdfs jstypez.Typedefs

		elemType := val.Type().Elem()
		elemZeroVal := reflect.Zero(elemType).Interface()

		elemTdf, eTdfs := p.ParseAny(ctx, baseId+"Element", elemZeroVal)
		checker.StrEmpty(elemTdf.Head.Name, "elemTdf.Head.Name")
		elemTdf.Head.Name = baseId + "Element"

		allTdfs.ExtractFrom(eTdfs)
		allTdfs.Add(elemTdf)

		return &jstypez.Typedef{
			Head: &jstypez.Tuple{Typ: fmt.Sprintf("Array<%s>", elemTdf.Head.Name)},
		}, &allTdfs

	case reflect.Map:
		var allTdfs jstypez.Typedefs

		mapType := val.Type()
		keyType := mapType.Key()
		keyZeroVal := reflect.Zero(keyType).Interface()

		elemType := mapType.Elem()
		elemZeroVal := reflect.Zero(elemType).Interface()

		keyTdf, kTdfs := p.ParseAny(ctx, baseId+"Key", keyZeroVal)
		valTdf, vTdfs := p.ParseAny(ctx, baseId+"Value", elemZeroVal)

		checker.StrEmpty(keyTdf.Head.Name, "keyTdf.Head.Name")
		checker.StrEmpty(valTdf.Head.Name, "valTdf.Head.Name")

		keyTdf.Head.Name = baseId + "Key"
		valTdf.Head.Name = baseId + "Value"

		allTdfs.ExtractFrom(kTdfs)
		allTdfs.ExtractFrom(vTdfs)
		allTdfs.Add(keyTdf, valTdf)

		return &jstypez.Typedef{
			Head: &jstypez.Tuple{Typ: fmt.Sprintf("Record<%s, %s>", keyTdf.Head.Name, valTdf.Head.Name)},
		}, &allTdfs

	default:
		panic(fmt.Sprintf("unsupported type %T / %s: %#v", v, val.Kind(), v))
	}
}

func (p *Parser) checkSchema(ctx context.Context, baseId string, val reflect.Value) (*jstypez.Typedef, *jstypez.Typedefs) {
	type S interface {
		Schema() *schemaz.Schema
	}
	v := val.Interface()
	var s *schemaz.Schema
	if v, ok := v.(S); ok {
		s = v.Schema()
	}
	rv := reflect.ValueOf(v)
	if s == nil && rv.IsValid() && rv.Kind() != reflect.Ptr {
		ptr := reflect.New(rv.Type())
		ptr.Elem().Set(rv)
		if ct, ok := ptr.Interface().(S); ok {
			s = ct.Schema()
		}
	}
	if s == nil {
		return nil, nil
	}
	a, b := p.ParseSchema(ctx, baseId, s)
	a.Head.Name = ""
	return a, b
}

func checkCustom(ctx context.Context, baseId string, val reflect.Value) (*jstypez.Typedef, *jstypez.Typedefs) {
	v := val.Interface()
	if ret, tdfs := Resolve(v); ret != nil {
		if ret.Head.Typ != jstypez.TypBoolean && ret.Head.Typ != jstypez.TypString && ret.Head.Typ != jstypez.TypNumber {
			logger.Info(ctx, "non basic parser custom detected", map[string]any{"typ": ret.Head.Typ, "base": baseId})
		}
		return ret, tdfs
	}
	return nil, nil
}
