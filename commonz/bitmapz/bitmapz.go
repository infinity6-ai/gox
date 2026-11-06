package bitmapz

import (
	"math/bits"
)

type Bits interface {
	~uint64
}

func ContainsAll[T Bits](current T, expected ...T) bool {
	all := Join(expected...)
	return current&all == all
}

func ContainsAny[T Bits](current T, expected ...T) bool {
	any := Join(expected...)
	return (current & any) != 0
}

func Split[T Bits](current T) []T {
	if current == 0 {
		return nil
	}
	count := bits.OnesCount64(uint64(current))
	result := make([]T, 0, count)
	for current != 0 {
		bit := current & -current
		result = append(result, bit)
		current &^= bit
	}
	return result
}

func Join[T Bits](values ...T) T {
	var result T
	for _, v := range values {
		result |= v
	}
	return result
}

func Set[T Bits](current T, enable bool, values ...T) T {
	mask := Join(values...)
	if enable {
		return current | mask
	}
	return current &^ mask
}
