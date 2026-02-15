package task

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
	HasTaskAck(key string) bool
	SetTaskAck(key string)
	HasTaskNack(key string) bool
	SetTaskNack(key string)
}

type taskRepository interface {
	GetProcessingWithID(ctx context.Context, id string) (*domain.Task, error)
	HasProcessingWithID(ctx context.Context, id string) (bool, error)
	Delete(ctx context.Context, id string) error
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

	err, rbErr := trx.Do(ctx, func(ctx context.Context) error {
		has, err := s.taskRepository.HasProcessingWithID(ctx, taskID)
		if err != nil {
			return fmt.Errorf("check processing task by id existence: %w", err)
		}

		if has {
			if err = s.taskRepository.Delete(ctx, taskID); err != nil {
				return fmt.Errorf("delete task: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		s.handleRollbackError("rollback failed on ack", rbErr, err)
		return err
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

	err, rbErr := trx.Do(ctx, func(ctx context.Context) error {
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
		s.handleRollbackError("rollback failed on nack", rbErr, err)
		return err
	}

	if idempotencyKey != nil {
		s.idempotencyKeyCache.SetTaskNack(*idempotencyKey)
	}
	return nil
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
