//go:build !testcover

package main

import (
	"fmt"

	"github.com/infinity6-ai/gox/noopz/version"
)

func main() {
	fmt.Println(version.Version())
}
