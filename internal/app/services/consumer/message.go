package consumer

import (
	"context"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/domain/consumer"
	"github.com/art-es/queue-service/internal/infra/log"
)

type messageHandler struct {
	queueService queueService
	taskService  taskService
	listenTasks  func(ctx context.Context, tasks <-chan *domain.Task, out chan<- *consumer.Message)
	logger       log.Logger
}

func newMessageHandler(
	queueService queueService,
	taskService taskService,
	logger log.Logger,
) *messageHandler {
	return &messageHandler{
		queueService: queueService,
		taskService:  taskService,
		listenTasks:  listenTasks,
		logger:       logger,
	}
}

func (h *messageHandler) handle(ctx context.Context, in *consumer.Message, out chan<- *consumer.Message) {
	switch in.Type {
	case consumer.InputTypeQueueSubscribe:
		h.handleQueueSubscribe(ctx, in, out)
	case consumer.InputTypeTaskAck:
		h.handleTaskAck(ctx, in, out)
	case consumer.InputTypeTaskNack:
		h.handleTaskNack(ctx, in, out)
	}
}

func (h *messageHandler) handleQueueSubscribe(ctx context.Context, in *consumer.Message, out chan<- *consumer.Message) {
	queueName, ok := in.Data.(consumer.MessageDataQueueName)
	if !ok {
		return
	}

	logger := h.logger.With("queue_name", string(queueName))

	tasks, err := h.queueService.Subscribe(ctx, string(queueName))
	if err != nil {
		logger.Log(log.LevelError).
			With("message", "queue subscribe error").
			With("error", err.Error()).
			Write()

		out <- &consumer.Message{
			Type: consumer.OutputTypeQueueSubscribePass,
		}
		return
	}

	logger.Log(log.LevelDebug).
		With("message", "subscribed to queue chan").
		Write()

	out <- &consumer.Message{
		Type: consumer.OutputTypeQueueSubscribeFail,
	}

	go h.listenTasks(ctx, tasks, out)
}

func (h *messageHandler) handleTaskAck(ctx context.Context, in *consumer.Message, out chan<- *consumer.Message) {
	taskID, ok := in.Data.(consumer.MessageDataTaskID)
	if !ok {
		return
	}

	if err := h.taskService.Ack(ctx, string(taskID)); err != nil {
		h.logger.Log(log.LevelError).
			With("message", "task ack error").
			With("task_id", string(taskID)).
			With("error", err.Error()).
			Write()

		out <- &consumer.Message{
			Type: consumer.OutputTypeTaskAckFail,
		}
		return
	}

	out <- &consumer.Message{
		Type: consumer.OutputTypeTaskAckPass,
	}
}

func (h *messageHandler) handleTaskNack(ctx context.Context, in *consumer.Message, out chan<- *consumer.Message) {
	taskID, ok := in.Data.(consumer.MessageDataTaskID)
	if !ok {
		return
	}

	if err := h.taskService.Nack(ctx, string(taskID)); err != nil {
		h.logger.Log(log.LevelError).
			With("message", "task nack error").
			With("task_id", string(taskID)).
			With("error", err.Error()).
			Write()

		out <- &consumer.Message{
			Type: consumer.OutputTypeTaskNackFail,
		}
		return
	}

	out <- &consumer.Message{
		Type: consumer.OutputTypeTaskNackPass,
	}
}
