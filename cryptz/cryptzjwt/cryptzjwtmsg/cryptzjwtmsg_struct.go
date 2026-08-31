package cryptzjwtmsg

import (
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/internal/pb/cryptzmsgpb"
)

type RSAMessageCiphered struct {
	Version        uint32
	Signature      []byte
	SrcKey         string
	DstKey         string
	CipheredSymKey []byte
	CipheredData   []byte
}

func (d *RSAMessageCiphered) Validate(SignatureSize int, CipheredSymKeySize int) {
	checker.Equal(uint32(1), d.Version, "Version")
	checker.Len(d.Signature, SignatureSize, "SignatureSize")
	checker.StrNotEmpty(d.SrcKey, "SrcKey")
	checker.StrNotEmpty(d.DstKey, "DstKey")
	checker.Len(d.CipheredSymKey, CipheredSymKeySize, "CipheredSymKeySize")
	checker.NotEmpty(d.CipheredData, "CipheredText")
}

func (d *RSAMessageCiphered) FromProto(pb *cryptzmsgpb.RSAMessageCiphered) {
	d.Version = pb.Version
	d.Signature = pb.Signature
	d.SrcKey = pb.SrcKey
	d.DstKey = pb.DstKey
	d.CipheredSymKey = pb.CipheredSymKey
	d.CipheredData = pb.CipheredData
}

func (d *RSAMessageCiphered) ToProto() *cryptzmsgpb.RSAMessageCiphered {
	return &cryptzmsgpb.RSAMessageCiphered{
		Version:        d.Version,
		Signature:      d.Signature,
		SrcKey:         d.SrcKey,
		DstKey:         d.DstKey,
		CipheredSymKey: d.CipheredSymKey,
		CipheredData:   d.CipheredData,
	}
}
