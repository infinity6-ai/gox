// Package constraintz provides a collection of generic type constraints for common Go types.
// These constraints are useful for writing generic functions that operate on specific
// categories of types, such as integers, floats, or numbers in general.
package constraintz

// SInts is a constraint that permits any signed integer type.
type SInts interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// UInts is a constraint that permits any unsigned integer type.
type UInts interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Ints is a constraint that permits any signed or unsigned integer type.
type Ints interface {
	SInts | UInts
}

// Floats is a constraint that permits any floating-point type.
type Floats interface {
	~float32 | ~float64
}

// Numbers is a constraint that permits any integer or floating-point type.
type Numbers interface {
	Ints | Floats
}

// Basic is a constraint that permits any string or numeric type.
type Basic interface {
	string | Numbers
}

// Void is a constraint that permits any type. It is an alias for `any`.
type Void any
