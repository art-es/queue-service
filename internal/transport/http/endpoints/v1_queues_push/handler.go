package v1_queues_push

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/art-es/queue-service/internal/app/services/queue"
	"github.com/art-es/queue-service/internal/infra/log"
	transport "github.com/art-es/queue-service/internal/transport/http"
)

const (
	maxPayloadSize = 1024 * 4
)

type requestBody struct {
	Payload string `json:"payload"`
}

type responseBody struct {
	Task *responseBodyTask `json:"task"`
}

type responseBodyTask struct {
	ID        string `json:"id"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}

type handler struct {
	queueService queueService
	logger       log.Logger
}

func newHandler(queueService queueService, logger log.Logger) *handler {
	logger = logger.With("module", "internal/transport/http/endpoints/v1_queues_push")

	return &handler{
		queueService: queueService,
		logger:       logger,
	}
}

func (h *handler) Handle(ctx transport.Context) {
	req := parseRequest(ctx)
	if req == nil {
		return
	}

	task, err := h.queueService.Push(ctx, req)
	if err != nil {
		h.logger.Log(log.LevelError).
			With("message", "queue service error").
			With("error", err.Error()).
			With("queue_name", req.QueueName).
			Write()

		transport.WriteInternalError(ctx)
		return
	}

	transport.Write(ctx, http.StatusCreated, &responseBody{
		Task: &responseBodyTask{
			ID:        task.ID,
			Payload:   task.Payload,
			CreatedAt: task.CreatedAt.Format(time.DateTime),
		},
	})
}

func parseRequest(ctx transport.Context) *queue.PushRequest {
	queueName := ctx.Request().PathValue("queueName")

	if len(queueName) == 0 {
		transport.WriteBadRequestFields(ctx, transport.CommonResponseBodyField{
			Name:   "queueName",
			Reason: transport.ReasonEmpty,
		})
		return nil
	}

	var rb requestBody
	if err := json.NewDecoder(ctx.Request().Body).Decode(&rb); err != nil {
		transport.WriteInvalidRequestBody(ctx)
		return nil
	}

	if rb.Payload == "" {
		transport.WriteBadRequestFields(ctx, transport.CommonResponseBodyField{
			Name:   "payload",
			Reason: transport.ReasonEmpty,
		})
		return nil
	}

	if len(rb.Payload) > maxPayloadSize {
		transport.WriteBadRequestFields(ctx, transport.CommonResponseBodyField{
			Name:   "payload",
			Reason: transport.ReasonTooLarge,
		})
		return nil
	}

	return &queue.PushRequest{
		IdempotencyKey: transport.GetIdempotencyKey(ctx),
		QueueName:      queueName,
		Payload:        rb.Payload,
	}
}
