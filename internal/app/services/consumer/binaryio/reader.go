package binary

import (
	"encoding/binary"
	"io"

	"github.com/art-es/queue-service/internal/app/services/consumer/dto"
)

const (
	inputTypeQueueSubscribe = uint8(dto.InputTypeQueueSubscribe)
	inputTypeTaskAck        = uint8(dto.InputTypeTaskAck)
	inputTypeTaskNack       = uint8(dto.InputTypeTaskNack)
)

type reader struct{}

func newReader() *reader {
	return &reader{}
}

func (*reader) Read(r io.Reader) (*dto.Message, error) {
	var msgType uint8
	if err := binary.Read(r, binary.BigEndian, &msgType); err != nil {
		return nil, err
	}

	var msgData any
	var err error

	switch msgType {
	case inputTypeQueueSubscribe:
		msgData, err = readQueueName(r)
	case inputTypeTaskAck, inputTypeTaskNack:
		msgData, err = readTaskID(r)
	default:
		// unsupported type
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &dto.Message{
		Type: dto.MessageType(msgType),
		Data: msgData,
	}, nil
}

func readQueueName(r io.Reader) (dto.MessageDataQueueName, error) {
	var val [256]byte
	if err := binary.Read(r, binary.BigEndian, &val); err != nil {
		return "", err
	}

	out := dto.MessageDataQueueName(convertBinaryQueueName(val))
	return out, nil
}

func readTaskID(r io.Reader) (dto.MessageDataTaskID, error) {
	var val [16]byte
	if err := binary.Read(r, binary.BigEndian, &val); err != nil {
		return "", err
	}

	out := dto.MessageDataTaskID(convertBinaryTaskID(val))
	return out, nil
}
