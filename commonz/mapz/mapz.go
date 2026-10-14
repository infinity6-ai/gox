package mapz

import (
	"maps"
	"slices"
)

func SortedKeys[M ~map[K]V, K string, V any](m M) []K {
	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)
	return keys
}
