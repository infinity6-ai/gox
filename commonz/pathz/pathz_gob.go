package pathz

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func (p *Path) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(p.String()); err != nil {
		return nil, fmt.Errorf("cannot marshal path to gob: %w", err)
	}
	return buf.Bytes(), nil
}

func (p *Path) GobDecode(data []byte) error {
	var s string
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(&s); err != nil {
		return fmt.Errorf("cannot unmarshal path from gob: %w", err)
	}

	err := p.Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse path from gob: %w", err)
	}
	return nil
}
