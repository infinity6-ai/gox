package supervalue

type checker interface {
	Check()
}

type X[T comparable] interface {
}

type SuperValue[T comparable] struct {
	v T
}

func Set[V comparable](sv *SuperValue[V], s any, v V) {
	sv.v = v
}

// func Set[V comparable](sv *SuperValue[V], s any, v V) {
// 	sv.v = v
// }

// func Set[T X[V], V comparable](sv T) {
// }

// func Set[T comparable](sv any, val any) {
// 	sv.(checker).Check()
// 	basePtr := SuperValue[T](sv)
// 	basePtr.v = val
// }

type MyValue SuperValue[string]

func (m MyValue) Check() {

}

func NewMyValue(val string) MyValue {
	ret := MyValue{}
	x := SuperValue[string](ret)
	Set(&x, ret, val)
	ret = MyValue(x)
	return ret
}
