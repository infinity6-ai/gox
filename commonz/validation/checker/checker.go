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

func Empty[S ~[]E, E any](actual S, msg string, args ...any) {
	errorz.Check(validation.Empty(actual, msg, args...))
}

func NotEmpty[S ~[]E, E any](actual S, msg string, args ...any) {
	errorz.Check(validation.NotEmpty(actual, msg, args...))
}

func StrContains(expected string, actual string, msg string, args ...any) {
	errorz.Check(validation.StrContains(expected, actual, msg, args...))
}

func StrEmpty(actual string, msg string, args ...any) {
	errorz.Check(validation.StrEmpty(actual, msg, args...))
}

func StrNotEmpty(actual string, msg string, args ...any) {
	errorz.Check(validation.StrNotEmpty(actual, msg, args...))
}

func StrNotContains(expected string, actual string, msg string, args ...any) {
	errorz.Check(validation.StrNotContains(expected, actual, msg, args...))
}

func True(actual bool, msg string, args ...any) {
	errorz.Check(validation.True(actual, msg, args...))
}

func False(actual bool, msg string, args ...any) {
	errorz.Check(validation.False(actual, msg, args...))
}

func StringRegex(pattern, value, msg string, args ...any) {
	errorz.Check(validation.StringRegex(pattern, value, msg, args...))
}

func NotNil(actual any, msg string, args ...any) {
	errorz.Check(validation.NotNil(actual, msg, args...))
}

func OneOf[T comparable](expected []T, actual T, msg string, args ...any) {
	errorz.Check(validation.OneOf(expected, actual, msg, args...))
}

func Len[S ~[]E, E any](actual S, length int, msg string, args ...any) {
	errorz.Check(validation.Len(actual, length, msg, args...))
}

