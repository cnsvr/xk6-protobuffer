package protobuf

import (
	"context"
	"fmt"

	"go.k6.io/k6/js/modules"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"github.com/bufbuild/protocompile"
)

func init() {
	modules.Register("k6/x/protobuffer", NewProtoBuffer())
}

type ProtoBuffer struct {
	Compiler *protocompile.Compiler
	Messages map[string]*ProtoMessage
}

func NewProtoBuffer() *ProtoBuffer {
	return &ProtoBuffer{
		Compiler: &protocompile.Compiler{
			Resolver: &protocompile.SourceResolver{},
		},
		Messages: make(map[string]*ProtoMessage),
	}
}

type ProtoMessage struct {
	MessageDesc protoreflect.MessageDescriptor
	Message     *dynamicpb.Message
}

func (p *ProtoBuffer) Load(protoFilePath, messageType string) (*ProtoMessage, error) {
	files, err := p.Compiler.Compile(context.Background(), protoFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto file: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files parsed in proto file: %s", protoFilePath)
	}

	messageDesc := files[0].Messages().ByName(protoreflect.Name(messageType))
	if messageDesc == nil {
		return nil, fmt.Errorf("message type '%s' not found in proto file '%s'", messageType, protoFilePath)
	}

	protoMessage := &ProtoMessage{
		MessageDesc: messageDesc,
		Message:     dynamicpb.NewMessage(messageDesc),
	}

	p.Messages[messageType] = protoMessage
	return protoMessage, nil
}

func (pm *ProtoMessage) Encode() ([]byte, error) {
	if pm.Message == nil {
		return nil, fmt.Errorf("no dynamic message to encode")
	}
	return proto.Marshal(pm.Message)
}

func (pm *ProtoMessage) Decode(protoData []byte) error {
	if pm.Message == nil {
		return fmt.Errorf("no dynamic message to decode into")
	}
	return proto.Unmarshal(protoData, pm.Message)
}

func (pm *ProtoMessage) SetField(field string, value interface{}) error {
	fieldDesc := pm.MessageDesc.Fields().ByName(protoreflect.Name(field))
	if fieldDesc == nil {
		return fmt.Errorf("field '%s' not found in message", field)
	}

	var protoValue protoreflect.Value
	switch fieldDesc.Kind() {
	case protoreflect.Int64Kind:
		switch v := value.(type) {
		case int:
			protoValue = protoreflect.ValueOf(int64(v))
		case int64:
			protoValue = protoreflect.ValueOf(v)
		case float64:
			protoValue = protoreflect.ValueOf(int64(v))
		default:
			return fmt.Errorf("field '%s' expects an int64-compatible value, got %T", field, value)
		}

	case protoreflect.Int32Kind:
		switch v := value.(type) {
		case int:
			protoValue = protoreflect.ValueOf(int32(v))
		case int32:
			protoValue = protoreflect.ValueOf(v)
		case float64:
			protoValue = protoreflect.ValueOf(int32(v))
		default:
			return fmt.Errorf("field '%s' expects an int32-compatible value, got %T", field, value)
		}

	case protoreflect.StringKind:
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("field '%s' expects a string value, got %T", field, value)
		}
		protoValue = protoreflect.ValueOf(strValue)

	default:
		return fmt.Errorf("unsupported field kind for '%s': %s", field, fieldDesc.Kind())
	}

	pm.Message.Set(fieldDesc, protoValue)
	return nil
}

func (pm *ProtoMessage) GetField(field string) (interface{}, error) {
	fieldDesc := pm.Message.Descriptor().Fields().ByName(protoreflect.Name(field))
	if fieldDesc == nil {
		return nil, fmt.Errorf("field '%s' not found in message", field)
	}
	return pm.Message.Get(fieldDesc).Interface(), nil
}
