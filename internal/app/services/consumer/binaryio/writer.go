package binary

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/art-es/queue-service/internal/app/domain/consumer"
)

type writer struct{}

func newWriter() *writer {
	return &writer{}
}

func (*writer) Write(w io.Writer, msg *consumer.Message) error {
	var (
		msgType     = uint8(msg.Type)
		msgData any = nil
	)

	switch msg.Data.(type) {
	case consumer.MessageDataTask:
		msgData = convertToBinaryTask(msg.Data.(consumer.MessageDataTask))
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
