package main

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func check2[T any](s T, err error) T {
	check(err)
	return s
}
