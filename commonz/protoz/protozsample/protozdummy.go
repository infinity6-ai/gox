package protozsample

import "github.com/infinity6-ai/gox/commonz/protoz/internal/pb/protozpb"

type Dummy struct {
	Id   uint32 `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func NewDummy(id uint32, name string) *Dummy {
	return &Dummy{Id: id, Name: name}
}

func (d *Dummy) FromProto(pb *protozpb.Dummy) {
	d.Id = pb.Id
	d.Name = pb.Name
}

func (d *Dummy) ToProto() *protozpb.Dummy {
	return &protozpb.Dummy{
		Id:   d.Id,
		Name: d.Name,
	}
}
