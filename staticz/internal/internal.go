package internal

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func Check2[T any](s T, err error) T {
	Check(err)
	return s
}
