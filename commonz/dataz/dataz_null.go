package dataz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// NullString represents a string that may be NULL.
type NullString struct {
	StringVal string
	Valid     bool
}

// MarshalJSON implements json.Marshaler.
func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.StringVal)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullString) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		n.StringVal = ""
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(b, &n.StringVal); err != nil {
		return fmt.Errorf("failed to unmarshal NullString: %w", err)
	}
	n.Valid = true
	return nil
}

// String returns the string representation.
func (n NullString) String() string {
	if !n.Valid {
		return "<null>"
	}
	return n.StringVal
}

// NullInt64 represents an int64 that may be NULL.
type NullInt64 struct {
	Int64 int64
	Valid bool
}

// MarshalJSON implements json.Marshaler.
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Int64)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullInt64) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		n.Int64 = 0
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(b, &n.Int64); err != nil {
		return fmt.Errorf("failed to unmarshal NullInt64: %w", err)
	}
	n.Valid = true
	return nil
}

// String returns the string representation.
func (n NullInt64) String() string {
	if !n.Valid {
		return "<null>"
	}
	return fmt.Sprintf("%d", n.Int64)
}

// NullFloat64 represents a float64 that may be NULL.
type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

// MarshalJSON implements json.Marshaler.
func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Float64)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullFloat64) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		n.Float64 = 0.0
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(b, &n.Float64); err != nil {
		return fmt.Errorf("failed to unmarshal NullFloat64: %w", err)
	}
	n.Valid = true
	return nil
}

// String returns the string representation.
func (n NullFloat64) String() string {
	if !n.Valid {
		return "<null>"
	}
	return fmt.Sprintf("%g", n.Float64)
}

// NullBool represents a bool that may be NULL.
type NullBool struct {
	Bool  bool
	Valid bool
}

// MarshalJSON implements json.Marshaler.
func (n NullBool) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Bool)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullBool) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		n.Bool = false
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(b, &n.Bool); err != nil {
		return fmt.Errorf("failed to unmarshal NullBool: %w", err)
	}
	n.Valid = true
	return nil
}

// String returns the string representation.
func (n NullBool) String() string {
	if !n.Valid {
		return "<null>"
	}
	return fmt.Sprintf("%t", n.Bool)
}

// NullTimestamp represents a time.Time that may be NULL.
type NullTimestamp struct {
	Timestamp time.Time
	Valid     bool
}

// MarshalJSON implements json.Marshaler.
func (n NullTimestamp) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Timestamp)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NullTimestamp) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		n.Timestamp = time.Time{}
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(b, &n.Timestamp); err != nil {
		return fmt.Errorf("failed to unmarshal NullTimestamp: %w", err)
	}
	n.Valid = true
	return nil
}

// String returns the string representation.
func (n NullTimestamp) String() string {
	if !n.Valid {
		return "<null>"
	}
	return n.Timestamp.Format(time.RFC3339Nano)
}
