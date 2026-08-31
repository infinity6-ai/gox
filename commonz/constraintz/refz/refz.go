package refz

func Point[T any](val T) *T {
	return &val
}
