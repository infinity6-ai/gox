package jsonz

import "go.code.infinity6.ai/platform/util/blobz"

func Parse[T any, I blobz.Data](data I) (*T, error) {
	panic("implement it")
}

func MustParse[T any, I blobz.Data](data I) *T {
	panic("implement it")
}

func Format[T any, O blobz.Data](data *T) (O, error) {
	panic("implement it")
}

func MustFormat[T any, O blobz.Data](data *T) O {
	panic("implement it")
}
