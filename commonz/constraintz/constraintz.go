// Package constraintz provides a collection of generic type constraints for common Go types.
// These constraints are useful for writing generic functions that operate on specific
// categories of types, such as integers, floats, or numbers in general.
package constraintz

import "golang.org/x/exp/constraints"

// Numbers is a constraint that permits any integer or floating-point type.
type Numbers interface {
	constraints.Integer | constraints.Float
}

// Basic is a constraint that permits any string or numeric type.
type Basic interface {
	string | Numbers
}

// Void is a constraint that permits any type. It is an alias for `any`.
type Void any
