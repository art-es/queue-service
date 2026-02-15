package v1_queues_push

import (
	"context"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/services/queue"
	"github.com/art-es/queue-service/internal/infra/log"
	transport "github.com/art-es/queue-service/internal/transport/http"
)

type queueService interface {
	Push(ctx context.Context, req *queue.PushRequest) (*domain.Task, error)
}

func Register(router transport.Router, queueService queueService, logger log.Logger) {
	router.Register("POST /v1/queues/{queueName}/push", newHandler(queueService, logger))
}
