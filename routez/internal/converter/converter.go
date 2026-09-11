package converter

import (
	"net/http"
	"strings"
	"unicode"
)

func Params2Json(params map[string]string) map[string][]string {
	n := make(map[string][]string, len(params))
	for k, v := range params {
		n[k] = []string{v}
	}
	return n
}

func Header2Json(headers http.Header) map[string][]string {
	n := make(map[string][]string, len(headers))
	for k, v := range headers {
		nk := strings.Map(func(r rune) rune {
			if r == '-' {
				return '_'
			}
			return unicode.ToLower(r)
		}, k)
		n[nk] = v
	}
	return n
}

func Json2Header(in map[string][]string, out http.Header) {
	for k, v := range in {
		nk := strings.ReplaceAll(k, "_", "-")
		out.Del(nk)
		for _, vv := range v {
			out.Add(nk, vv)
		}
	}
}

func Json2Params(in map[string][]string) map[string]string {
	n := make(map[string]string, len(in))
	for k, v := range in {
		if len(v) > 0 {
			n[k] = v[0]
		}
	}
	return n
}
