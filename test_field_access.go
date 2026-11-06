package main

import "fmt"

type MyConstraint interface {
	~struct {
		Value uint64
		Name  string
	}
}

type Bit struct {
	Value uint64
	Name  string
}

func PrintValue[T MyConstraint](v T) {
	b := Bit(v)
	fmt.Println(b.Value)
}

func main() {
	type MyStruct struct {
		Value uint64
		Name  string
	}
	PrintValue(MyStruct{Value: 42, Name: "test"})
}
