package v1_queues_pop

import (
	"net/http"
	"time"

	"github.com/art-es/queue-service/internal/infra/log"
	transport "github.com/art-es/queue-service/internal/transport/http"
)

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
	logger = logger.With("module", "internal/transport/http/endpoints/v1_queues_pop")

	return &handler{
		queueService: queueService,
		logger:       logger,
	}
}

func (h *handler) Handle(ctx transport.Context) {
	queueName := ctx.Request().PathValue("queueName")

	if len(queueName) == 0 {
		transport.WriteBadRequestFields(ctx, transport.CommonResponseBodyField{
			Name:   "queueName",
			Reason: transport.ReasonEmpty,
		})
		return
	}

	task, err := h.queueService.Pop(ctx, queueName)
	if err != nil {
		h.logger.Log(log.LevelError).
			With("message", "queue service error").
			With("error", err.Error()).
			With("queue_name", queueName).
			Write()

		transport.WriteInternalError(ctx)
		return
	}

	if task == nil {
		transport.WriteEmpty(ctx, http.StatusNoContent)
		return
	}

	transport.Write(ctx, http.StatusOK, &responseBody{
		Task: &responseBodyTask{
			ID:        task.ID,
			Payload:   task.Payload,
			CreatedAt: task.CreatedAt.Format(time.DateTime),
		},
	})
}
