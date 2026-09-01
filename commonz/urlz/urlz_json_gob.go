package urlz

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

func (u *Url) GobEncode() ([]byte, error) {
	s, err := u.ToString()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(s); err != nil {
		return nil, fmt.Errorf("cannot marshal url to gob: %w", err)
	}
	return buf.Bytes(), nil
}

func (u *Url) GobDecode(data []byte) error {
	var s string
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(&s); err != nil {
		return fmt.Errorf("cannot unmarshal url from gob: %w", err)
	}

	parsedUrl, err := Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse url from gob: %w", err)
	}

	*u = *parsedUrl
	return nil
}

func (u *Url) MarshalJSON() ([]byte, error) {
	s, err := u.ToString()
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

func (u *Url) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal url from json: %w", err)
	}
	if len(s) == 0 {
		return nil
	}
	parsedUrl, err := Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse url from json: %w", err)
	}
	*u = *parsedUrl
	return nil
}
