package binary

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/art-es/queue-service/internal/app/services/consumer/dto"
)

type writer struct{}

func newWriter() *writer {
	return &writer{}
}

func (*writer) Write(w io.Writer, msg *dto.Message) error {
	var (
		msgType     = uint8(msg.Type)
		msgData any = nil
	)

	switch msg.Data.(type) {
	case dto.MessageDataTask:
		msgData = convertToBinaryTask(msg.Data.(dto.MessageDataTask))
	}

	if err := binary.Write(w, binary.BigEndian, msgType); err != nil {
		return fmt.Errorf("write message type: %w", err)
	}
	if msgData != nil {
		if err := binary.Write(w, binary.BigEndian, msgData); err != nil {
			return fmt.Errorf("write message data: %w", err)
		}
	}
	return nil
}
