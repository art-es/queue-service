package dto

import "time"

const (
	InputTypeQueueSubscribe MessageType = iota + 1
	InputTypeTaskAck
	InputTypeTaskNack
)

const (
	OutputTypeQueueSubscribePass MessageType = iota + 1
	OutputTypeQueueSubscribeFail
	OutputTypeTaskAckPass
	OutputTypeTaskAckFail
	OutputTypeTaskNackPass
	OutputTypeTaskNackFail
	OutputTypeTaskProcess
)

type Message struct {
	Type MessageType
	Data any
}

type MessageType uint8

type (
	MessageDataQueueName string
	MessageDataTaskID    string
	MessageDataTask      struct {
		ID        string
		Payload   string
		CreatedAt time.Time
	}
)
