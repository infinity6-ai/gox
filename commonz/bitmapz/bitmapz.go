package bitmapz

type Bit interface {
	Name() string
}

type Bits[T Bit] interface {
	Value() uint64
}

type bitImpl struct {
	v uint64
	n string
}

func (x bitImpl) Name() string {
	return x.n
}

func create[T Bit](i uint64, s string) T {
	var ret Bit
	ret = &bitImpl{v: i, n: s}
	return ret.(T)
}

type Grant Bit

var Read Grant = create[Grant](1, "READ")

type Grants uint64

func (v Grants) Value() uint64 { return uint64(v) }

// func removeit() {
// 	x := Grants(1)
// 	ContainsAll(x, Read)
// 	var s []Grant
// 	Split(x, &s)
// }

func ContainsAll[T Bits[B], B Bit](a T, b B) bool {
	panic("aaa")
}

func Split[T Bits[B], B Bit](current T, target *[]B) []B {
	panic("xx")
}

// implement: ContainsAny, Split, Join
