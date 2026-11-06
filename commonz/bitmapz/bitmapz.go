package bitmapz

import "reflect"

// 1. Define the memory layout constraint for our generics
type BitConstraint interface {
	~struct {
		Value uint64
		Name  string
	}
}

// 2. The base struct
type Bit struct {
	Value uint64
	Name  string
}

// 3. The generic Bitmap type. It's a uint64 tied to a specific BitConstraint.
type Bitmap[T BitConstraint] uint64

// --- Registry for Value Discovery ---

var registry = make(map[reflect.Type]any)

func Register[T BitConstraint](bits ...T) {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	var existing []T
	if v, ok := registry[typ]; ok {
		existing = v.([]T)
	}
	registry[typ] = append(existing, bits...)
}

// 4. Expand implementation.
// Notice it takes `n T` (the uint64 mask) and returns `[]B` (the struct slices).
func Expand[B BitConstraint, T ~uint64](n T) []B {
	typ := reflect.TypeOf((*B)(nil)).Elem()
	known, ok := registry[typ].([]B)
	if !ok {
		return nil
	}

	var result []B
	mask := uint64(n)
	for _, b := range known {
		bit := Bit(b)
		if bit.Value == 0 {
			if mask == 0 {
				result = append(result, b)
			}
			continue
		}
		if mask&bit.Value == bit.Value {
			result = append(result, b)
		}
	}
	return result
}
