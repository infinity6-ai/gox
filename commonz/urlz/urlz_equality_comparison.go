package urlz

import "cmp"

// Equals checks if two Url objects are equal.
// Two URLs are considered equal if all their components are identical.
func (u *Url) Equals(other *Url) bool {
	if u == nil && other == nil {
		return true
	}
	if u == nil || other == nil {
		return false
	}

	return u.Scheme == other.Scheme &&
		u.User == other.User &&
		u.Password == other.Password &&
		u.Host == other.Host &&
		u.Port == other.Port &&
		u.Path.Equals(other.Path) &&
		u.Query == other.Query &&
		u.Fragment == other.Fragment
}

// Compare compares two Url objects based on their string representation.
// It returns:
// - 0 if the URLs are equal (u.String() == other.String())
// - -1 if u's string representation is lexicographically smaller than other's
// - 1 if u's string representation is lexicographically greater than other's
func (u *Url) Compare(other *Url) int {
	return cmp.Compare(u.String(), other.String())
}
