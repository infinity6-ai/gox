package cryptzsalt

import (
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptzhash"
	"github.com/infinity6-ai/gox/cryptz/cryptzrand"
)

const SaltSize = 16
const MaxSize = 512 * 1024

type Salted struct {
	Salt   []byte
	Result []byte
}

func (s *Salted) BundleSize() int {
	return 2 + len(s.Salt) + len(s.Result)
}

func (s *Salted) Format() []byte {
	checker.Equal(SaltSize, len(s.Salt), "s.Salt.length")
	checker.LessOrEqual(len(s.Result), MaxSize, "s.Result.length")

	total := uint16(s.BundleSize() - 2)

	ret := make([]byte, 0, 2+total)
	ret = binary.BigEndian.AppendUint16(ret, total)
	ret = append(ret, s.Salt...)
	ret = append(ret, s.Result...)
	return ret
}

func Generate[T blobz.Data](original T) (*Salted, error) {
	originalBytes := blobz.ToBytes(original)
	salt := cryptzrand.Rand(SaltSize)
	return compute(salt, originalBytes)
}

func compute(salt []byte, originalBytes []byte) (*Salted, error) {
	if len(salt) != SaltSize {
		return nil, fmt.Errorf("invalid salt size: expected %d, got %d", SaltSize, len(salt))
	}
	if len(originalBytes) > MaxSize {
		return nil, fmt.Errorf("original data size %d exceeds max size %d", len(originalBytes), MaxSize)
	}
	secret, err := cryptzhash.SHA256Data(append(slices.Clone(salt), originalBytes...))
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}
	return &Salted{
		Salt:   salt,
		Result: secret,
	}, nil
}

func Verify[T blobz.Data](bundle []byte, original T) (*Salted, error) {
	if len(bundle) < 2 {
		return nil, errors.New("bundle too short")
	}

	total := binary.BigEndian.Uint16(bundle[:2])
	if int(total) > MaxSize {
		return nil, fmt.Errorf("bundle total size %d exceeds max size %d", total, MaxSize)
	}
	if len(bundle) < int(total)+2 {
		return nil, fmt.Errorf("bundle length %d is smaller than declared total %d", len(bundle), total+2)
	}

	bundle = bundle[2 : total+2]
	if len(bundle) < SaltSize {
		return nil, errors.New("bundle does not contain a full salt")
	}
	salt := bundle[:SaltSize]
	expectedHash := bundle[SaltSize:]

	expected, err := compute(salt, blobz.ToBytes(original))
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash for verification: %w", err)
	}

	if subtle.ConstantTimeCompare(expected.Result, expectedHash) != 1 {
		return nil, errors.New("hash mismatch")
	}

	expected.Result = expectedHash
	return expected, nil
}
