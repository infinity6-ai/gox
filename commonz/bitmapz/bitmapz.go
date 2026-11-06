package bitmapz

import (
	"math/bits"

	"golang.org/x/exp/constraints"
)

func ContainsAll[T constraints.Unsigned](current T, expected ...T) bool {
	all := Join(expected...)
	return current&all == all
}

func ContainsAny[T constraints.Unsigned](current T, expected ...T) bool {
	any := Join(expected...)
	return (current & any) != 0
}

func Split[T constraints.Unsigned](current T) []T {
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

func Join[T constraints.Unsigned](values ...T) T {
	var result T
	for _, v := range values {
		result |= v
	}
	return result
}

func Set[T constraints.Unsigned](current T, enable bool, values ...T) T {
	mask := Join(values...)
	if enable {
		return current | mask
	}
	return current &^ mask
}
