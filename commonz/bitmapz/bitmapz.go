package bitmapz

import (
	"math/bits"
)

// Bits is a constraint that permits any unsigned integer type used for bitmasks.
type Bits interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// ContainsAll reports whether current contains all bits in expected.
// If expected is empty, it returns true.
func ContainsAll[T Bits](current T, expected ...T) bool {
	all := Join(expected...)
	return current&all == all
}

// ContainsAny reports whether current contains at least one of the bits in expected.
// If expected is empty or current is 0, it returns false.
func ContainsAny[T Bits](current T, expected ...T) bool {
	any := Join(expected...)
	return (current & any) != 0
}

// Split decomposes current into a slice of individual set bits in ascending order.
// If current is 0, it returns nil.
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

// Join combines all given values into a single bitmask using bitwise OR.
func Join[T Bits](values ...T) T {
	var result T
	for _, v := range values {
		result |= v
	}
	return result
}

// Set adds or removes the given values from current based on enable.
func Set[T Bits](current T, enable bool, values ...T) T {
	mask := Join(values...)
	if enable {
		return current | mask
	}
	return current &^ mask
}
