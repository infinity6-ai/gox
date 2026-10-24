package cmdcryptz

import (
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt"
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

func parseIat(iatStr string) (time.Time, error) {
	if iatStr == "" {
		return time.Now().UTC(), nil
	}
	if n, err := strconv.ParseInt(iatStr, 10, 64); err == nil {
		if n < 10000000000 {
			return time.Unix(n, 0).UTC(), nil
		}
		return time.UnixMilli(n).UTC(), nil
	}
	if d, err := time.ParseDuration(iatStr); err == nil {
		if d > 0 {
			d = -d
		}
		return time.Now().UTC().Add(d), nil
	}
	if t, err := time.Parse(time.RFC3339, iatStr); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", iatStr); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("invalid --iat format %q: expected unix epoch (seconds or ms), RFC3339, or relative offset (e.g. -5m)", iatStr)
}

func parseBxt(bxtRaw, bxtHex, bxtB64, bxtB64Url, bxtFile string) ([]byte, error) {
	if bxtFile != "" {
		data, err := os.ReadFile(bxtFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read bxt file %s: %w", bxtFile, err)
		}
		return data, nil
	}
	if bxtHex != "" {
		data, err := hex.DecodeString(cleanHexInput([]byte(bxtHex)))
		if err != nil {
			return nil, fmt.Errorf("invalid --bxt-hex: %w", err)
		}
		return data, nil
	}
	if bxtB64 != "" {
		blob, err := cryptzb64.StdDecode(cleanHexInput([]byte(bxtB64)))
		if err != nil {
			return nil, fmt.Errorf("invalid --bxt-b64: %w", err)
		}
		return blob.Bytes(), nil
	}
	if bxtB64Url != "" {
		blob, err := cryptzb64.UrlDecode(cleanHexInput([]byte(bxtB64Url)))
		if err != nil {
			return nil, fmt.Errorf("invalid --bxt-b64url: %w", err)
		}
		return blob.Bytes(), nil
	}
	if bxtRaw != "" {
		return []byte(bxtRaw), nil
	}
	return nil, nil
}

func buildJwtPayload(jti, typ string, iatTime time.Time, expDuration time.Duration, iss, aud, sub, ext string, bxtBytes []byte, idp string) (*cryptzjwt.JWTPayload, error) {
	if jti != "" {
		step := cryptzjwt.NewToken().
			SetJti(jti).
			SetPayloadType(typ).
			CreatedAt(iatTime).
			ExpiresIn(expDuration).
			SetIssuer(iss).
			SetAudience(aud).
			SetSubject(sub)

		if len(bxtBytes) > 0 && ext != "" {
			return step.SetBothExtentions(ext, bxtBytes).SetIdentityProvider(idp).BuildPayload()
		} else if len(bxtBytes) > 0 {
			return step.SetBinaryExtension(bxtBytes).SetIdentityProvider(idp).BuildPayload()
		}
		return step.SetExtension(ext).SetIdentityProvider(idp).BuildPayload()
	}

	step := cryptzjwt.NewToken().
		GenerateJti().
		SetPayloadType(typ).
		CreatedAt(iatTime).
		ExpiresIn(expDuration).
		SetIssuer(iss).
		SetAudience(aud).
		SetSubject(sub)

	if len(bxtBytes) > 0 && ext != "" {
		return step.SetBothExtentions(ext, bxtBytes).SetIdentityProvider(idp).BuildPayload()
	} else if len(bxtBytes) > 0 {
		return step.SetBinaryExtension(bxtBytes).SetIdentityProvider(idp).BuildPayload()
	}
	return step.SetExtension(ext).SetIdentityProvider(idp).BuildPayload()
}
