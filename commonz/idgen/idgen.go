package idgen

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/infinity6-ai/gox/cryptz/cryptzb32"
)

func Bytes() []byte {
	ret := uuid.New()
	return ret[:]
}

func String() string {
	return uuid.New().String()
}

func FromString(str string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(str)).String()
}

func B64() string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(Bytes())
}

func B32() string {
	return cryptzb32.Encode(Bytes()).String()
}

func Hex() string {
	return hex.EncodeToString(Bytes())
}

func UInt() (uint64, uint64) {
	data := Bytes()
	high := binary.BigEndian.Uint64(data[:8])
	low := binary.BigEndian.Uint64(data[8:])
	return high, low
}

func Int() (int64, int64) {
	high, low := UInt()
	return int64(high), int64(low)
}
