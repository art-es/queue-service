package v1_tasks_ack

import (
	"context"

	"github.com/art-es/queue-service/internal/infra/log"
	transport "github.com/art-es/queue-service/internal/transport/http"
)

type taskService interface {
	Ack(ctx context.Context, taskID string, idempotencyKey *string) error
}

func Register(router transport.Router, taskService taskService, logger log.Logger) {
	router.Register("POST /v1/tasks/{taskId}/ack", newHandler(taskService, logger))
}
