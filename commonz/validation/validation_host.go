package validation

import (
	"net"
	"regexp"
)

var hostnameRegex = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)

func Host(host string, msg string, args ...any) error {
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}
	if !hostnameRegex.MatchString(host) {
		return newError("invalid hostname", map[string]any{"host": host}, msg, args...)
	}
	return nil
}
