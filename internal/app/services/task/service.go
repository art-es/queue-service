package task

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
	HasTaskAck(key string) bool
	SetTaskAck(key string)
	HasTaskNack(key string) bool
	SetTaskNack(key string)
}

type taskRepository interface {
	Complete(ctx context.Context, id string) error
	GetProcessingWithID(ctx context.Context, id string) (*domain.Task, error)
	Save(ctx context.Context, task *domain.Task) error
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
	return &Service{
		clock:               clock,
		idempotencyKeyCache: idempotencyKeyCache,
		taskRepository:      taskRepository,
		logger:              logger,
	}
}

func (s *Service) Ack(ctx context.Context, taskID string, idempotencyKey *string) error {
	if idempotencyKey != nil {
		if s.idempotencyKeyCache.HasTaskAck(*idempotencyKey) {
			return nil
		}
	}

	if err := s.taskRepository.Complete(ctx, taskID); err != nil {
		return nil
	}

	if idempotencyKey != nil {
		s.idempotencyKeyCache.SetTaskAck(*idempotencyKey)
	}
	return nil
}

func (s *Service) Nack(ctx context.Context, taskID string, idempotencyKey *string) error {
	if idempotencyKey != nil {
		if s.idempotencyKeyCache.HasTaskNack(*idempotencyKey) {
			return nil
		}
	}

	now := s.clock.Now()
	err := trxutil.DoOrLogError(s.logger, "task.nack", ctx, func(ctx context.Context) error {
		task, err := s.taskRepository.GetProcessingWithID(ctx, taskID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil
			}

			return fmt.Errorf("check processing task by id existence: %w", err)
		}

		task.ToFailed(now)

		if err = s.taskRepository.Save(ctx, task); err != nil {
			return fmt.Errorf("save task: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if idempotencyKey != nil {
		s.idempotencyKeyCache.SetTaskNack(*idempotencyKey)
	}
	return nil
}
