package urlz

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
)

// http(s)://aaa....
// file:/tmp/x , file:///tmp/x
// unix:/tmp/socket
// gs://bla/ble

type Url struct {
	Scheme   string
	User     string
	Password string
	Host     string
	Port     string
	Path     *pathz.Path
	Query    string
	Fragment string
}

func Parse(urlStr string) (*Url, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("error parseing url: %s, %w", urlStr, err)
	}
	provider, err := getProvider(u.Scheme)
	if err != nil {
		return nil, err
	}
	ret, err := provider.Parser(u)
	if err != nil {
		return nil, err
	}
	err = provider.Validation(ret)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func (u *Url) ToString() (string, error) {
	provider, err := getProvider(u.Scheme)
	if err != nil {
		return "", err
	}
	err = provider.Validation(u)
	if err != nil {
		return "", err
	}
	return provider.ToString(u), nil
}

func (u *Url) String() string {
	s, err := u.ToString()
	errorz.Check(err)
	return s
}

func (u *Url) Clone() *Url {
	return &Url{
		Scheme:   u.Scheme,
		User:     u.User,
		Password: u.Password,
		Host:     u.Host,
		Port:     u.Port,
		Path:     u.Path.Clone(),
		Query:    u.Query,
		Fragment: u.Fragment,
	}
}

// IsBaseOf checks if the current URL (u) is a base for the provided 'other' URL.
// It is considered a base if:
// 1. The Scheme, User, Password, Host, and Port are identical.
// 2. The Path of 'u' is a base of 'other.Path' (as determined by pathz.IsBaseOf).
// Query and Fragment are not considered in this check.
func (u *Url) IsBaseOf(other *Url) bool {
	if u.Scheme != other.Scheme {
		return false
	}
	if u.User != other.User {
		return false
	}
	if u.Password != other.Password {
		return false
	}
	if u.Host != other.Host {
		return false
	}
	if u.Port != other.Port {
		return false
	}
	if u.Path == nil || other.Path == nil {
		return u.Path == other.Path
	}
	return u.Path.IsBaseOf(other.Path)
}

func (u *Url) JoinPath(others ...*pathz.Path) (*Url, error) {
	p, err := u.Path.Join(others...)
	if err != nil {
		err = fmt.Errorf("%w: error joining path to url %s: %s", err, u, others)
	}
	var ret *Url
	if p != nil {
		ret = u.Clone()
		ret.Path = p
	}
	return ret, err
}

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
