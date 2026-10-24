package cmdcryptz

import (
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaservice"
	"github.com/spf13/cobra"
)

func readInput(cmd *cobra.Command, args []string) ([]byte, error) {
	if len(args) > 0 {
		return []byte(strings.Join(args, " ")), nil
	}
	in := cmd.InOrStdin()
	if f, ok := in.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return nil, err
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, fmt.Errorf("no input provided")
		}
	}
	data, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no input provided")
	}
	return data, nil
}

func cleanHexInput(in []byte) string {
	return strings.TrimSpace(string(in))
}

func loadKey(keyFlag string) (string, error) {
	if keyFlag == "" {
		return "", fmt.Errorf("key must be specified")
	}
	if keyFlag == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read key from stdin: %w", err)
		}
		return string(data), nil
	}
	if _, err := os.Stat(keyFlag); err == nil {
		data, err := os.ReadFile(keyFlag)
		if err != nil {
			return "", fmt.Errorf("failed to read key file %s: %w", keyFlag, err)
		}
		return string(data), nil
	}
	return keyFlag, nil
}

func outputBytes(cmd *cobra.Command, data []byte, b64 bool, b64url bool, raw bool) error {
	out := cmd.OutOrStdout()
	if raw {
		_, err := out.Write(data)
		return err
	}
	if b64url {
		fmt.Fprintln(out, cryptzb64.UrlEncode(data).String())
		return nil
	}
	if b64 {
		fmt.Fprintln(out, cryptzb64.StdEncode(data).String())
		return nil
	}
	fmt.Fprintln(out, hex.EncodeToString(data))
	return nil
}

func decodeInputBytes(data []byte, b64 bool, b64url bool) ([]byte, error) {
	str := cleanHexInput(data)
	if b64url {
		blob, err := cryptzb64.UrlDecode(str)
		if err != nil {
			return nil, fmt.Errorf("invalid b64url: %w", err)
		}
		return blob.Bytes(), nil
	}
	if b64 {
		blob, err := cryptzb64.StdDecode(str)
		if err != nil {
			return nil, fmt.Errorf("invalid b64: %w", err)
		}
		return blob.Bytes(), nil
	}
	decoded, err := hex.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	return decoded, nil
}

func safePrivKeyImport(s cryptzrsaservice.RSAService, keyPEM string) (priv *rsa.PrivateKey, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to import private key: %v", r)
		}
	}()
	return s.PrivKeyImport(keyPEM), nil
}

func safePubKeyImport(s cryptzrsaservice.RSAService, keyPEM string) (pub *rsa.PublicKey, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to import public key: %v", r)
		}
	}()
	return s.PubKeyImport(keyPEM), nil
}
