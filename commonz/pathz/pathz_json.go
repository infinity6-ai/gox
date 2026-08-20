package pathz

import (
	"encoding/json"
	"fmt"
)

func (p *Path) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Path) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal path from json: %w", err)
	}

	parsed, err := Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse path from json: %w", err)
	}

	p.Parts = parsed.Parts
	p.Parents = parsed.Parents
	p.HasEndingSlash = parsed.HasEndingSlash
	return nil
}
