package configz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
)

var ErrNotConfigured = errors.New("unknown config")

type Prop string

var props = map[Prop]string{}

type reverter map[Prop]string

func (p reverter) Close() error {
	maps.Copy(props, p)
	return nil
}

func Create(name Prop, defval string) Prop {
	_, ok := props[name]
	if ok {
		panic(fmt.Errorf("property already defined: %s", name))
	}
	ret := name
	v := os.Getenv(string(name))
	if v == "" {
		v = defval
	}
	props[name] = v
	return ret
}

func CreateEncoded(name Prop, defval string) Prop {
	return Create(name, cryptzb64.UrlEncode(defval).String())
}

func (p Prop) Get(ctx context.Context) (string, error) {
	ret, ok := props[p]
	if !ok {
		return "", fmt.Errorf("property does not exist: %s, %w", p, ErrNotConfigured)
	}
	return ret, nil
}

func (p Prop) Decoded(ctx context.Context) (string, error) {
	ret, err := p.Get(ctx)
	if err != nil {
		return "", err
	}
	if ret != "" {
		dec, err := cryptzb64.UrlDecode(ret)
		if err != nil {
			return "", err
		}
		ret = dec.String()
	}
	return ret, nil
}

func (p Prop) Req(ctx context.Context) string {
	ret, err := p.Get(ctx)
	errorz.Check(err)
	if ret == "" {
		panic(fmt.Errorf("property is required: %s", p))
	}
	return ret
}

func (p Prop) ReqDecoded(ctx context.Context) string {
	ret, err := p.Decoded(ctx)
	errorz.Check(err)
	if ret == "" {
		panic(fmt.Errorf("property is required: %s", p))
	}
	return ret
}

func (p Prop) Set(ctx context.Context, value string, values ...any) (io.Closer, error) {
	return Sets(ctx, map[Prop]string{p: fmt.Sprintf(value, values...)})
}

func (p Prop) SetEncoded(ctx context.Context, value string, values ...any) (io.Closer, error) {
	encoded := cryptzb64.UrlEncode(fmt.Sprintf(value, values...)).String()
	return p.Set(ctx, "%s", encoded)
}

func Sets(ctx context.Context, values map[Prop]string) (io.Closer, error) {
	ret := reverter{}
	for k, v := range values {
		old, ok := props[k]
		if !ok {
			return nil, fmt.Errorf("property is required: %s", k)
		}
		props[k] = v
		ret[k] = old
	}
	return ret, nil
}
