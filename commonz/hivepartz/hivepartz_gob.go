package hivepartz

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func (p *HiveParts) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(p.FormatString()); err != nil {
		return nil, fmt.Errorf("cannot marshal path to gob: %w", err)
	}
	return buf.Bytes(), nil
}

func (p *HiveParts) GobDecode(data []byte) error {
	var s string
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(&s); err != nil {
		return fmt.Errorf("cannot unmarshal path from gob: %w", err)
	}

	np, r, err := ParseString(s)
	if err != nil {
		return fmt.Errorf("cannot parse path from gob: %w", err)
	}
	if r.Parents() != 0 || r.HasEndingSlash() || r.PartsLen() != 0 {
		return fmt.Errorf("path unsupported: %s, remaning: %s", s, r)
	}
	p.names = np.names
	p.values = np.values
	return nil
}
