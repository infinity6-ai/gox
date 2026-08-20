package jsonz

import "github.com/infinity6-ai/gox/commonz/constraintz"

func Parse[T any, I constraintz.Data](data I) (*T, error) {
	panic("implement it")
}

func MustParse[T any, I constraintz.Data](data I) *T {
	panic("implement it")
}

func Format[T any, O constraintz.Data](data *T) (O, error) {
	panic("implement it")
}

func MustFormat[T any, O constraintz.Data](data *T) O {
	panic("implement it")
}
