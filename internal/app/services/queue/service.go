//go:generate mockgen -source=service.go -destination=service_mock_test.go -package=$GOPACKAGE
package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/repository"
	"github.com/art-es/queue-service/internal/infra/log"
	"github.com/art-es/queue-service/internal/infra/trx/trxutil"
)

type clock interface {
	Now() time.Time
}

type idempotencyKeyCache interface {
	GetQueuePush(key string) (*domain.Task, bool)
	SetQueuePush(key string, result *domain.Task)
}

type taskRepository interface {
	GetFirstPending(ctx context.Context, queueName string) (*domain.Task, error)
	Save(ctx context.Context, task *domain.Task) error
}

type PushRequest struct {
	QueueName      string
	Payload        string
	IdempotencyKey *string
}

type Service struct {
	clock               clock
	idempotencyKeyCache idempotencyKeyCache
	taskRepository      taskRepository
	logger              log.Logger
}

func NewService(
	clock clock,
	idempotencyKeyCache idempotencyKeyCache,
	taskRepository taskRepository,
	logger log.Logger,
) *Service {
	logger = logger.With("module", "internal/app/services/queue")

	return &Service{
		clock:               clock,
		idempotencyKeyCache: idempotencyKeyCache,
		taskRepository:      taskRepository,
		logger:              logger,
	}
}

func (s *Service) Push(ctx context.Context, req *PushRequest) (*domain.Task, error) {
	if req.IdempotencyKey != nil {
		if task, ok := s.idempotencyKeyCache.GetQueuePush(*req.IdempotencyKey); ok {
			return task, nil
		}
	}

	task := domain.NewTask(req.QueueName, req.Payload)

	if err := s.taskRepository.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("save task: %w", err)
	}

	if req.IdempotencyKey != nil {
		s.idempotencyKeyCache.SetQueuePush(*req.IdempotencyKey, task)
	}

	return task, nil
}

func (s *Service) Pop(ctx context.Context, queueName string) (*domain.Task, error) {
	var task *domain.Task

	now := s.clock.Now()
	err := trxutil.DoOrLogError(s.logger, "queue.pop", ctx, func(ctx context.Context) error {
		var err error

		task, err = s.taskRepository.GetFirstPending(ctx, queueName)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil
			}

			return fmt.Errorf("get first pending task: %w", err)
		}

		task.ToProcessing(now)

		if err = s.taskRepository.Save(ctx, task); err != nil {
			return fmt.Errorf("save task: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}
