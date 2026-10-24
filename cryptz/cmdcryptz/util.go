package cmdcryptz

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func readInput(args []string) ([]byte, error) {
	if len(args) > 0 {
		return []byte(strings.Join(args, " ")), nil
	}
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return io.ReadAll(os.Stdin)
	}
	return nil, fmt.Errorf("no input provided")
}

func cleanHexInput(in []byte) string {
	return strings.TrimSpace(string(in))
}
