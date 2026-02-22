package binary

import (
	"github.com/art-es/queue-service/internal/app/services/consumer/dto"
)

const (
	sizeUUID      = 16
	sizeDateTime  = 19
	sizeShortText = 256
	sizeLongText  = 1024
)

var (
	convertBinaryQueueName = convertShortBytesToString
	convertBinaryTaskID    = convertUUIDBytesToString
)

type messageDataTask struct {
	ID        [sizeUUID]byte
	Payload   [sizeLongText]byte
	CreatedAt [sizeDateTime]byte
}

func convertToBinaryTask(task dto.MessageDataTask) messageDataTask {
	return messageDataTask{
		ID:        convertStringToUUIDBytes(task.ID),
		Payload:   convertStringToLongBytes(task.Payload),
		CreatedAt: convertTimeToDateTimeBytes(task.CreatedAt),
	}
}
