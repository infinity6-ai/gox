package cryptzjwtstate

import (
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/cryptz/cryptzjwt/internal/pb/cryptzstatepb"
)

type StateMessageCiphered struct {
	Version      uint32
	SrcKey       string
	DstKey       string
	CipheredData []byte
	KeyId        string
}

func (d *StateMessageCiphered) Validate() {
	checker.Equal(uint32(1), d.Version, "Version")
	checker.StrNotEmpty(d.SrcKey, "SrcKey")
	checker.StrNotEmpty(d.DstKey, "DstKey")
	checker.NotEmpty(d.CipheredData, "CipheredData")
}

func (d *StateMessageCiphered) FromProto(pb *cryptzstatepb.StateMessageCiphered) {
	d.Version = pb.Version
	d.SrcKey = pb.SrcKey
	d.DstKey = pb.DstKey
	d.CipheredData = pb.CipheredData
	d.KeyId = pb.KeyId
}

func (d *StateMessageCiphered) ToProto() *cryptzstatepb.StateMessageCiphered {
	return &cryptzstatepb.StateMessageCiphered{
		Version:      d.Version,
		SrcKey:       d.SrcKey,
		DstKey:       d.DstKey,
		CipheredData: d.CipheredData,
		KeyId:        d.KeyId,
	}
}
