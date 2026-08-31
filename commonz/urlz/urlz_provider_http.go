package urlz

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation"
)

func (p *providerSpec) providerHttp() (*providerSpec, error) {
	p.Validation = func(u *Url) error {
		err := validation.OneOf([]string{"http", "https"}, u.Scheme, "scheme")
		if err != nil {
			return err
		}
		err = validation.StrNotEmpty(u.Host, "host")
		if err != nil {
			return err
		}

		trimmedHost := strings.TrimPrefix(strings.TrimSuffix(u.Host, "]"), "[")
		if ip := net.ParseIP(trimmedHost); ip == nil { // Not an IP, so it must be a hostname
			// This regex is a simplified version for common hostnames, excluding IPs and special characters not allowed in domain labels.
			// It allows alphanumeric characters and hyphens, but not leading/trailing hyphens.
			// It does not enforce TLD rules, as that's complex and often changes.
			hostnameRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}\.?$`)
			if !hostnameRegex.MatchString(u.Host) {
				return fmt.Errorf("validation error invalid hostname: %s", u.Host)
			}
		}

		// Path can be empty, but if it is not, it must be absolute
		if u.Path != nil && (len(u.Path.Parts()) > 0 || u.Path.HasEndingSlash()) {
			err = u.Path.ValidateAbsoluteFile()
			if err != nil {
				return err
			}
		}

		return nil
	}
	p.Parser = func(u *url.URL) (*Url, error) {
		pt, err := pathz.Parse(u.Path)
		if err != nil {
			return nil, fmt.Errorf("error parsing http url path: %s, %w", u.Path, err)
		}
		password, _ := u.User.Password()
		return &Url{
			Scheme:   u.Scheme,
			User:     u.User.Username(),
			Password: password,
			Host:     u.Hostname(),
			Port:     u.Port(),
			Path:     pt,
			Query:    u.RawQuery,
			Fragment: u.Fragment,
		}, nil
	}
	p.ToString = func(u *Url) string {
		var user *url.Userinfo
		if u.User != "" {
			if u.Password != "" {
				user = url.UserPassword(u.User, u.Password)
			} else {
				user = url.User(u.User)
			}
		}

		host := u.Host
		ip := net.ParseIP(host)
		if ip != nil && ip.To4() == nil {
			host = fmt.Sprintf("[%s]", host)
		}

		if u.Port != "" {
			host = fmt.Sprintf("%s:%s", host, u.Port)
		}

		ret := &url.URL{
			Scheme:   u.Scheme,
			User:     user,
			Host:     host,
			Path:     u.Path.String(),
			RawQuery: u.Query,
			Fragment: u.Fragment,
		}
		return ret.String()
	}
	return p, nil
}
