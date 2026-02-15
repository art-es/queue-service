package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/repository"
	"github.com/art-es/queue-service/internal/infra/log"
	"github.com/art-es/queue-service/internal/infra/ops"
	"github.com/art-es/queue-service/internal/infra/trx"
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
	IdempotencyKey *string
	QueueName      string
	Payload        string
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
		task, ok := s.idempotencyKeyCache.GetQueuePush(*req.IdempotencyKey)
		if !ok {
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

	err, rbErr := trx.Do(ctx, func(ctx context.Context) error {
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
		s.handleRollbackError("rollback failed on pop", rbErr, err)
		return nil, err
	}

	return task, nil
}

func (s *Service) handleRollbackError(msg string, rbErr error, opErr error) {
	if rbErr != nil {
		s.logger.Log(log.LevelError).
			With("message", msg).
			With("rb_error", rbErr.Error()).
			With("op_error", ops.ErrorMessage(opErr)).
			Write()
	}
}
