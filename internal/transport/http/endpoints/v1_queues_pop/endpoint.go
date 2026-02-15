package v1_queues_pop

import (
	"context"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/infra/log"
	transport "github.com/art-es/queue-service/internal/transport/http"
)

type queueService interface {
	Pop(ctx context.Context, queueName string) (*domain.Task, error)
}

func Register(router transport.Router, queueService queueService, logger log.Logger) {
	router.Register("POST /v1/queues/{queueName}/pop", newHandler(queueService, logger))
}
