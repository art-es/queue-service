package binary

import (
	"encoding/binary"
	"io"

	"github.com/art-es/queue-service/internal/app/domain/consumer"
)

const (
	inputTypeQueueSubscribe = uint8(consumer.InputTypeQueueSubscribe)
	inputTypeTaskAck        = uint8(consumer.InputTypeTaskAck)
	inputTypeTaskNack       = uint8(consumer.InputTypeTaskNack)
)

type reader struct{}

func newReader() *reader {
	return &reader{}
}

func (*reader) Read(r io.Reader) (*consumer.Message, error) {
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
		// TODO: add log
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &consumer.Message{
		Type: consumer.MessageType(msgType),
		Data: msgData,
	}, nil
}

func readQueueName(r io.Reader) (consumer.MessageDataQueueName, error) {
	var val [256]byte
	if err := binary.Read(r, binary.BigEndian, &val); err != nil {
		return "", err
	}

	out := consumer.MessageDataQueueName(convertBinaryQueueName(val))
	return out, nil
}

func readTaskID(r io.Reader) (consumer.MessageDataTaskID, error) {
	var val [16]byte
	if err := binary.Read(r, binary.BigEndian, &val); err != nil {
		return "", err
	}

	out := consumer.MessageDataTaskID(convertBinaryTaskID(val))
	return out, nil
}
