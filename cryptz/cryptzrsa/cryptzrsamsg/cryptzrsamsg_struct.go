package cryptzrsamsg

import (
	"github.com/infinity6-ai/gox/cryptz/pb/cryptzmsgpb"
	"go.code.infinity6.ai/platform/validation"
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
	validation.Equal(uint32(1), d.Version, "Version")
	validation.Len(d.Signature, SignatureSize, "SignatureSize")
	validation.NotEmpty(d.SrcKey, "SrcKey")
	validation.NotEmpty(d.DstKey, "DstKey")
	validation.Len(d.CipheredSymKey, CipheredSymKeySize, "CipheredSymKeySize")
	validation.NotEmpty(d.CipheredData, "CipheredText")
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
