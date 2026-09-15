package hivepartz

import (
	"encoding/json"
	"fmt"
)

func (p *HiveParts) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.FormatString())
}

func (p *HiveParts) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal path from json: %w", err)
	}
	if len(s) == 0 {
		return nil
	}
	np, r, err := ParseString(s)
	if err != nil {
		return fmt.Errorf("cannot parse path from json: %w", err)
	}
	if r.Parents() != 0 || r.HasEndingSlash() || r.PartsLen() != 0 {
		return fmt.Errorf("path unsupported: %s, remaning: %s", s, r)
	}
	p.names = np.names
	p.values = np.values
	return nil
}
