package protoz

import (
	"bytes"
	"io"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/proto"
)

type ProtoWrapper[T proto.Message] interface {
	FromProto(pb T)
	ToProto() T
}

func Parse[T ProtoWrapper[M], M proto.Message](data []byte, v T) T {
	msg := any(v).(ProtoWrapper[M])
	pb := msg.ToProto()
	err := proto.Unmarshal(data, pb)
	errorz.Check(err)
	msg.FromProto(pb)
	return msg.(T)
}

func Format[T ProtoWrapper[M], M proto.Message](v T) []byte {
	msg := any(v).(ProtoWrapper[M])
	pb := msg.ToProto()
	ret, err := proto.Marshal(pb)
	errorz.Check(err)
	return ret
}

func FormatSlice[T ProtoWrapper[M], M proto.Message](msgs []T) []byte {
	var buf bytes.Buffer
	for _, m := range msgs {
		msg := any(m).(ProtoWrapper[M])
		pb := msg.ToProto()
		_, err := protodelim.MarshalTo(&buf, pb)
		errorz.Check(err)
	}
	return buf.Bytes()
}

func ParseSlice[T ProtoWrapper[M], M proto.Message](data []byte) []T {
	var ret []T
	reader := bytes.NewReader(data)
	var tmp T
	typ := reflect.TypeOf(tmp).Elem()
	for {
		t := reflect.New(typ).Interface().(T)
		msg := any(t).(ProtoWrapper[M])
		pb := msg.ToProto()
		err := protodelim.UnmarshalFrom(reader, pb)
		if err == io.EOF {
			break
		}
		errorz.Check(err)
		msg.FromProto(pb)
		ret = append(ret, t)
	}
	return ret
}
