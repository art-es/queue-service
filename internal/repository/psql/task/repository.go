package queue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/repository"
	"github.com/art-es/queue-service/internal/infra/ops"
	"github.com/art-es/queue-service/internal/repository/psql"
)

type Repository struct {
	execGetter psql.ExecGetter
}

func NewRepository(execGetter psql.ExecGetter) *Repository {
	return &Repository{execGetter: execGetter}
}

func (r *Repository) GetFirstPending(ctx context.Context, queueName string) (*domain.Task, error) {
	exec, err := r.execGetter.Get(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, queue_name, payload, status, created_at, locked_until, last_fail_duration
		FROM tasks t
		WHERE t.queue_name = $1 AND t.status = 'pending'
		ORDER BY t.created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED`

	return getTask(exec, ctx, query, []any{queueName})
}

func (r *Repository) GetProcessingWithID(ctx context.Context, id string) (*domain.Task, error) {
	exec, err := r.execGetter.Get(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, queue_name, payload, status, created_at, locked_until, last_fail_duration
		FROM tasks t
		WHERE t.id = $1 AND t.status = 'processing'
		FOR UPDATE SKIP LOCKED`

	return getTask(exec, ctx, query, []any{id})
}

func (r *Repository) HasProcessingWithID(ctx context.Context, id string) (bool, error) {
	exec, err := r.execGetter.Get(ctx)
	if err != nil {
		return false, err
	}

	query := `
		SELECT id
		FROM tasks t
		WHERE t.id = $1 AND t.status = 'processing'
		FOR UPDATE SKIP LOCKED`

	var foundID int
	if err := exec.QueryRow(ctx, query, id).Scan(&foundID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("execute sql query: %w", err)
	}

	return true, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	exec, err := r.execGetter.Get(ctx)
	if err != nil {
		return err
	}

	query := "DELETE FROM tasks WHERE id = $1"

	if _, err = exec.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("execute sql query: %w", err)
	}

	return nil
}

func (r *Repository) Save(ctx context.Context, task *domain.Task) error {
	if task.ID == "" {
		return r.insert(ctx, task)
	}

	return r.update(ctx, task)
}

func (r *Repository) insert(ctx context.Context, task *domain.Task) error {
	exec, err := r.execGetter.Get(ctx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO tasks (queue_name, payload, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	args := []any{task.QueueName, task.Payload, task.Status}

	if err = exec.QueryRow(ctx, query, args...).Scan(&task.ID, &task.CreatedAt); err != nil {
		return fmt.Errorf("execute sql query: %w", err)
	}

	return nil
}

func (r *Repository) update(ctx context.Context, task *domain.Task) error {
	exec, err := r.execGetter.Get(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE tasks
		SET status = $2, locked_until = $3, last_fail_duration = $4
		WHERE id = $1`
	args := []any{task.ID, task.Status, task.LockedUntil, toSQLDuration(task.LastFailDuration)}

	if _, err = exec.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("execute sql query: %w", err)
	}

	return nil
}

func getTask(
	exec psql.Executer,
	ctx context.Context,
	query string,
	args []any,
) (*domain.Task, error) {
	task := &domain.Task{}
	lastFailDuration := sql.NullInt64{}
	scanDest := []any{
		&task.ID,
		&task.QueueName,
		&task.Payload,
		&task.Status,
		&task.CreatedAt,
		&task.LockedUntil,
		&lastFailDuration,
	}

	if err := exec.QueryRow(ctx, query, args...).Scan(scanDest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("execute sql query: %w", err)
	}

	task.LastFailDuration = fromSQLDuration(lastFailDuration)
	return task, nil
}

func toSQLDuration(in *time.Duration) *int64 {
	if in != nil {
		return ops.Pointer(int64(*in))
	}
	return nil
}

func fromSQLDuration(in sql.NullInt64) *time.Duration {
	if in.Valid {
		return ops.Pointer(time.Duration(in.Int64))
	}
	return nil
}
