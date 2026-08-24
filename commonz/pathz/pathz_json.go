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
	if len(s) == 0 {
		return nil
	}
	err := p.Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse path from json: %w", err)
	}
	return nil
}
