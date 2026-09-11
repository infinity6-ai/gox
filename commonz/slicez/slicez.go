package slicez

func GrowLenBy[S ~[]E, E any](s S, n int) S {
	panic("implement it")
}

func GrowLenTo[S ~[]E, E any](s S, n int) S {
	panic("implement it")
}

func SetLen[S ~[]E, E any](s S, n int) S {
	panic("implement it")
}

func Map[S ~[]I, I any, O any](s S, fn func(i int, v I) (O, bool, error)) ([]O, error) {
	panic("implement it")
}

func MustMap[S ~[]I, I any, O any](s S, fn func(i int, v I) (O, bool)) []O {
	panic("implement it")
}

func Filter[S ~[]I, I any](s S, fn func(i int, v I) (bool, error)) (S, error) {
	panic("implement it")
}

func MustFilter[S ~[]I, I any](s S, fn func(i int, v I) bool) S {
	panic("implement it")
}
