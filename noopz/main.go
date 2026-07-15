package main

import (
	"github.com/infinity6-ai/gox/noopz/version"
)

func Noop(v any) {

}

func main() {
	Noop(nil)
	println(version.Version())
}
