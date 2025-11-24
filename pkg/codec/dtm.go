package codec

import (
	"fmt"
	"go-micro.dev/v4/codec/bytes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type DtmCodec struct{}

func (d *DtmCodec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case *bytes.Frame:
		return m.Data, nil
	case proto.Message:
		return proto.Marshal(m)
	case protoiface.MessageV1:
		// #2333 compatible with etcd legacy proto.Message
		m2 := protoimpl.X.ProtoMessageV2Of(m)
		return proto.Marshal(m2)
	}
	return nil, fmt.Errorf("failed to marshal with dtm: %v is not type of *bytes.Frame or proto.Message", v)
}

func (d *DtmCodec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case proto.Message:
		return proto.Unmarshal(data, m)
	case protoiface.MessageV1:
		// #2333 compatible with etcd legacy proto.Message
		m2 := protoimpl.X.ProtoMessageV2Of(m)
		return proto.Unmarshal(data, m2)
	}
	return fmt.Errorf("failed to unmarshal with dtm: %v is not type of proto.Message", v)
}

func (d *DtmCodec) Name() string {
	return "grpc+dtm_raw"
}

func NewDtmCodec() encoding.Codec {
	return &DtmCodec{}
}
