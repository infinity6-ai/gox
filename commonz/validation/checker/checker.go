package checker

import (
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation"
	"golang.org/x/exp/constraints"
)

func Equal[T comparable](expected T, actual T, msg string, args ...any) {
	errorz.Check(validation.Equal(expected, actual, msg, args...))
}

func NotEqual[T comparable](expected T, actual T, msg string, args ...any) {
	errorz.Check(validation.NotEqual(expected, actual, msg, args...))
}

func Greater[T constraints.Ordered](value, threshold T, msg string, args ...any) {
	errorz.Check(validation.Greater(value, threshold, msg, args...))
}

func GreaterOrEqual[T constraints.Ordered](value, threshold T, msg string, args ...any) {
	errorz.Check(validation.GreaterOrEqual(value, threshold, msg, args...))
}

func Less[T constraints.Ordered](value, threshold T, msg string, args ...any) {
	errorz.Check(validation.Less(value, threshold, msg, args...))
}

func LessOrEqual[T constraints.Ordered](value, threshold T, msg string, args ...any) {
	errorz.Check(validation.LessOrEqual(value, threshold, msg, args...))
}

func Fail(msg string, args ...any) {
	errorz.Check(validation.Fail(msg, args...))
}
