package v1_tasks_nack

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/art-es/queue-service/internal/infra/log"
	transport "github.com/art-es/queue-service/internal/transport/http"
)

type handler struct {
	taskService taskService
	logger      log.Logger
}

func newHandler(taskService taskService, logger log.Logger) *handler {
	logger = logger.With("module", "internal/transport/http/endpoints/v1_tasks_nack")

	return &handler{
		taskService: taskService,
		logger:      logger,
	}
}

func (h *handler) Handle(ctx transport.Context) {
	taskID := ctx.Request().PathValue("taskId")

	if len(taskID) == 0 {
		transport.WriteBadRequestFields(ctx, transport.CommonResponseBodyField{
			Name:   "taskId",
			Reason: transport.ReasonEmpty,
		})
		return
	}

	if err := uuid.Validate(taskID); err != nil {
		transport.WriteBadRequestFields(ctx, transport.CommonResponseBodyField{
			Name:    "taskId",
			Reason:  transport.ReasonInvalid,
			Message: err.Error(),
		})
		return
	}

	if err := h.taskService.Nack(ctx, taskID, transport.GetIdempotencyKey(ctx)); err != nil {
		h.logger.Log(log.LevelError).
			With("message", "task service error").
			With("error", err.Error()).
			With("task_id", taskID).
			Write()

		transport.WriteInternalError(ctx)
		return
	}

	transport.WriteEmpty(ctx, http.StatusNoContent)
}
