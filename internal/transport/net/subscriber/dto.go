package subscriber

const (
	inTypeQueueSubscribe uint8 = iota + 1
	inTypeTaskAck
	inTypeTaskNack
	inTypeConnClose
)

const (
	outTypeQueueSubscribePass uint8 = iota + 1
	outTypeQueueSubscribeFail
	outTypeTaskAckPass
	outTypeTaskAckFail
	outTypeTaskNackPass
	outTypeTaskNackFail
	outTypeTaskProcess
)

const (
	sizeQueueName     = 256
	sizeTaskID        = 16
	sizeTaskPayload   = 1024
	sizeTaskCreatedAt = 19
)

type processTaskOut struct {
	ID        [sizeTaskID]byte
	Payload   [sizeTaskPayload]byte
	CreatedAt [sizeTaskCreatedAt]byte
}
