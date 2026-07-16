package dataz

import (
	"github.com/infinity6-ai/gox/commonz/constraintz"
)

func Limited[T constraintz.Data](s T, max int) T {
	if len(s) <= max {
		return s
	}
	return s[0:max]
}
