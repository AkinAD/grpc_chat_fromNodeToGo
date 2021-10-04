package models

import (
	manager "github.com/AkinAD/grpc_chat_fromNodeToGo/models/Manager"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func DoProtoMarshal(item protoreflect.ProtoMessage, logger manager.Logger) []byte{
	m := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: true,
	}

	json, err := m.Marshal(item)
	if err != nil {
		logger.Error.Panicf("ProtoJson unnable to Marshal value: %v", err)
	}
	return json
}