package bitmapz

import (
	"fmt"
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

// Serialize encodes the bitmask v into a compact variable-length byte slice.
// If v fits in 4 bits (0..15), it is serialized in 1 byte:
// the upper 4 bits are 0000, and the lower 4 bits store the value.
// Otherwise, the upper 4 bits specify the number of following payload bytes (1..8),
// the lower 4 bits are 0000, and the payload bytes follow in little-endian order.
// The total length ranges from 1 byte (for values 0..15) up to 9 bytes (1 header + 8 bytes for 64-bit).
func Serialize[T Bits](v T) []byte {
	u := uint64(v)
	if u <= 0x0F {
		return []byte{byte(u)}
	}

	size := byte((bits.Len64(u) + 7) / 8)
	buf := make([]byte, 1+size)
	buf[0] = size << 4

	for i := byte(0); i < size; i++ {
		buf[1+i] = byte(u >> (8 * i))
	}
	return buf
}

// Deserialize decodes a bitmask from data according to the variable-length encoding.
// It returns an error if data is empty, has an invalid length, invalid size, non-zero
// lower nibble when size > 0, or if the decoded value overflows type T.
func Deserialize[T Bits](data []byte) (T, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("data is empty")
	}

	header := data[0]
	size := header >> 4

	if size > 8 {
		return 0, fmt.Errorf("invalid size %d: bitmask size cannot exceed 8 bytes", size)
	}

	if size == 0 {
		if len(data) != 1 {
			return 0, fmt.Errorf("invalid data length: expected 1 byte for 4-bit value, got %d", len(data))
		}
		return T(header & 0x0F), nil
	}

	if header&0x0F != 0 {
		return 0, fmt.Errorf("invalid header 0x%02x: lower 4 bits must be 0 when size > 0", header)
	}

	expectedLen := 1 + int(size)
	if len(data) != expectedLen {
		return 0, fmt.Errorf("invalid data length: expected %d bytes for size %d, got %d", expectedLen, size, len(data))
	}

	var val uint64
	for i := 0; i < int(size); i++ {
		val |= uint64(data[1+i]) << (8 * i)
	}

	res := T(val)
	if uint64(res) != val {
		return 0, fmt.Errorf("value %d overflows type %T", val, res)
	}

	return res, nil
}
