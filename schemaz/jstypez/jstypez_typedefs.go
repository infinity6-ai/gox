package jstypez

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

type Tuple struct {
	Typ     string `json:"typ"`
	GoTyp   string `json:"go_typ,omitempty"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

type Tuples struct {
	m map[string]Tuple
	s []string
}

func NewTuples(tuples []Tuple) *Tuples {
	ret := &Tuples{}
	ret.Add(tuples...)
	return ret
}

func (p *Tuples) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.All())
}

func (p *Tuples) UnmarshalJSON(data []byte) error {
	var s []Tuple
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal path from json: %w", err)
	}
	p.Add(s...)
	return nil
}

func (t Tuples) All() []Tuple {
	ret := make([]Tuple, len(t.s))
	for idx, name := range t.s {
		ret[idx] = t.m[name]
	}
	return ret
}

func (t *Tuples) Add(tuples ...Tuple) {
	// Lazy initialization of the map
	if t.m == nil {
		t.m = make(map[string]Tuple)
	}

	for _, n := range tuples {
		old, ok := t.m[n.Name]
		if ok {
			if old != n {
				panic(fmt.Sprintf("collision, old: %#v, new: %#v", old, n))
			}
			// If it already exists and is identical, skip appending to avoid slice duplicates
			continue
		}

		t.m[n.Name] = n
		t.s = append(t.s, n.Name)
	}
}

type Typedef struct {
	Head       *Tuple  `json:"head"`
	Properties *Tuples `json:"properties"`
}

type Typedefs struct {
	m map[string]*Typedef
}

func (t *Typedefs) Merge(other Typedefs) {
	if t.m == nil {
		t.m = make(map[string]*Typedef)
	}
	for k, v := range other.m {
		if _, ok := t.m[k]; !ok {
			t.m[k] = v
		}
	}
}

func NewTypedefs(tuples []*Typedef) Typedefs {
	ret := Typedefs{}
	ret.Add(tuples...)
	return ret
}

func (p *Typedefs) ExtractFrom(tds *Typedefs) {
	if tds != nil {
		p.Add(tds.All()...)
	}
}

func (p *Typedefs) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.m)
}

func (p *Typedefs) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &p.m); err != nil {
		return fmt.Errorf("cannot unmarshal path from json: %w", err)
	}
	return nil
}

func (t *Typedefs) All() []*Typedef {
	ret := slices.Collect(maps.Values(t.m))
	slices.SortFunc(ret, func(a, b *Typedef) int {
		return strings.Compare(a.Head.Name, b.Head.Name)
	})
	return ret
}

func (t *Typedefs) Add(typedefs ...*Typedef) {
	// Lazy initialization of the map
	if t.m == nil {
		t.m = make(map[string]*Typedef)
	}

	for _, n := range typedefs {
		if n.Head.Name == "" {
			panic(fmt.Sprintf("name is required: %#v", n))
		}
		// Using Head.Name as the unique key for the typedef
		old, ok := t.m[n.Head.Name]
		if ok {
			// Typedef contains uncomparable fields (map and slice inside Tuples),
			// so we must use reflect.DeepEqual to check if they are identical.
			if !reflect.DeepEqual(old, n) {
				panic(fmt.Sprintf("collision, old: %#v, new: %#v", old, n))
			}
			// If it already exists and is completely identical, skip adding
			continue
		}

		t.m[n.Head.Name] = n
	}
}

type CustomType interface {
	JSTypedef() (*Typedef, *Typedefs)
}

type CustomTypeSimple interface {
	JSType() string
}
