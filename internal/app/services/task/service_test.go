package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/repository"
	"github.com/art-es/queue-service/internal/infra/log"
	"github.com/art-es/queue-service/internal/infra/log/logimpl"
	"github.com/art-es/queue-service/internal/infra/ops"
	"github.com/art-es/queue-service/internal/infra/trx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestService_Ack(t *testing.T) {
	var (
		ctx            = context.Background()
		taskID         = "testTaskID"
		idempotencyKey = "testIdempotencyKey"
	)

	type testDeps struct {
		mockIdempotencyKeyCache *MockidempotencyKeyCache
		mockTaskRepository      *MocktaskRepository
		service                 *Service
	}

	for _, tc := range []struct {
		name string
		run  func(t *testing.T, d testDeps)
	}{
		{
			name: "ack task",
			run: func(t *testing.T, d testDeps) {
				d.mockTaskRepository.EXPECT().
					Complete(gomock.Any(), gomock.Eq(taskID)).
					Return(nil)

				err := d.service.Ack(ctx, taskID, nil)
				assert.NoError(t, err)
			},
		},
		{
			name: "ack task with idempotency key",
			run: func(t *testing.T, d testDeps) {
				d.mockIdempotencyKeyCache.EXPECT().
					HasTaskAck(gomock.Eq(idempotencyKey)).
					Return(false)

				d.mockTaskRepository.EXPECT().
					Complete(gomock.Any(), gomock.Eq(taskID)).
					Return(nil)

				d.mockIdempotencyKeyCache.EXPECT().
					SetTaskAck(gomock.Eq(idempotencyKey))

				err := d.service.Ack(ctx, taskID, ops.Pointer(idempotencyKey))
				assert.NoError(t, err)
			},
		},
		{
			name: "idempotency key exists in cache",
			run: func(t *testing.T, d testDeps) {
				d.mockIdempotencyKeyCache.EXPECT().
					HasTaskAck(gomock.Eq(idempotencyKey)).
					Return(true)

				err := d.service.Ack(ctx, taskID, ops.Pointer(idempotencyKey))
				assert.NoError(t, err)
			},
		},
		{
			name: "complete task error",
			run: func(t *testing.T, d testDeps) {
				d.mockTaskRepository.EXPECT().
					Complete(gomock.Any(), gomock.Eq(taskID)).
					Return(errors.New("test error"))

				err := d.service.Ack(ctx, taskID, nil)
				assert.EqualError(t, err, "complete task: test error")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()

			mockIdempotencyKeyCache := NewMockidempotencyKeyCache(mc)
			mockTaskRepository := NewMocktaskRepository(mc)
			logger, _ := logimpl.NewTestLogger()

			tc.run(t, testDeps{
				mockIdempotencyKeyCache: mockIdempotencyKeyCache,
				mockTaskRepository:      mockTaskRepository,
				service:                 NewService(nil, mockIdempotencyKeyCache, mockTaskRepository, logger),
			})
		})
	}
}

func TestService_Nack(t *testing.T) {
	var (
		ctx            = context.Background()
		taskID         = "testTaskID"
		idempotencyKey = "testIdempotencyKey"
	)

	type testDeps struct {
		mockClock               *Mockclock
		mockIdempotencyKeyCache *MockidempotencyKeyCache
		mockTaskRepository      *MocktaskRepository
		logbuf                  log.Buffer
		service                 *Service
	}

	for _, tc := range []struct {
		name string
		run  func(t *testing.T, d testDeps)
	}{
		{
			name: "nack task first time",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:05:05")

				expTaskFromRepo := &domain.Task{
					ID:          taskID,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:08:08")),
				}

				expTaskAfterTransition := &domain.Task{
					ID:               taskID,
					Status:           domain.TaskStatusFailed,
					LockedUntil:      ops.Pointer(getTime(t, "2006-01-02 15:06:05")),
					LastFailDuration: ops.Pointer(time.Minute),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetProcessingWithID(gomock.Any(), gomock.Eq(taskID)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(expTaskFromRepo, nil)

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTaskAfterTransition)).
					Do(func(ctx context.Context, _ *domain.Task) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(nil)

				err := d.service.Nack(ctx, taskID, nil)
				assert.NoError(t, err)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "nack task not first time",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:05:05")

				expTaskFromRepo := &domain.Task{
					ID:               taskID,
					Status:           domain.TaskStatusProcessing,
					LockedUntil:      ops.Pointer(getTime(t, "2006-01-02 15:08:08")),
					LastFailDuration: ops.Pointer(4 * time.Minute),
				}

				expTaskAfterTransition := &domain.Task{
					ID:               taskID,
					Status:           domain.TaskStatusFailed,
					LockedUntil:      ops.Pointer(getTime(t, "2006-01-02 15:13:05")),
					LastFailDuration: ops.Pointer(8 * time.Minute),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetProcessingWithID(gomock.Any(), gomock.Eq(taskID)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(expTaskFromRepo, nil)

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTaskAfterTransition)).
					Do(func(ctx context.Context, _ *domain.Task) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(nil)

				err := d.service.Nack(ctx, taskID, nil)
				assert.NoError(t, err)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "nack task with idempotency key",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:05:05")

				expTaskFromRepo := &domain.Task{
					ID:          taskID,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:08:08")),
				}

				expTaskAfterTransition := &domain.Task{
					ID:               taskID,
					Status:           domain.TaskStatusFailed,
					LockedUntil:      ops.Pointer(getTime(t, "2006-01-02 15:06:05")),
					LastFailDuration: ops.Pointer(time.Minute),
				}

				d.mockIdempotencyKeyCache.EXPECT().
					HasTaskNack(gomock.Eq(idempotencyKey)).
					Return(false)

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetProcessingWithID(gomock.Any(), gomock.Eq(taskID)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(expTaskFromRepo, nil)

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTaskAfterTransition)).
					Do(func(ctx context.Context, _ *domain.Task) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(nil)

				d.mockIdempotencyKeyCache.EXPECT().
					SetTaskNack(gomock.Eq(idempotencyKey))

				err := d.service.Nack(ctx, taskID, ops.Pointer(idempotencyKey))
				assert.NoError(t, err)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "idempotency key exists in cache",
			run: func(t *testing.T, d testDeps) {
				d.mockIdempotencyKeyCache.EXPECT().
					HasTaskNack(gomock.Eq(idempotencyKey)).
					Return(true)

				err := d.service.Nack(ctx, taskID, ops.Pointer(idempotencyKey))
				assert.NoError(t, err)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "commit trx error",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:05:05")

				expTaskFromRepo := &domain.Task{
					ID:          taskID,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:08:08")),
				}

				expTaskAfterTransition := &domain.Task{
					ID:               taskID,
					Status:           domain.TaskStatusFailed,
					LockedUntil:      ops.Pointer(getTime(t, "2006-01-02 15:06:05")),
					LastFailDuration: ops.Pointer(time.Minute),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetProcessingWithID(gomock.Any(), gomock.Eq(taskID)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(expTaskFromRepo, nil)

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTaskAfterTransition)).
					Do(func(ctx context.Context, _ *domain.Task) {
						assert.True(t, trx.Exists(ctx), "transaction exists")

						trx.AddCommit(ctx, func() error {
							return errors.New("test error")
						})
					}).
					Return(nil)

				err := d.service.Nack(ctx, taskID, nil)
				assert.EqualError(t, err, "commit trx: test error")
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "save task error",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:05:05")

				expTaskFromRepo := &domain.Task{
					ID:          taskID,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:08:08")),
				}

				expTaskAfterTransition := &domain.Task{
					ID:               taskID,
					Status:           domain.TaskStatusFailed,
					LockedUntil:      ops.Pointer(getTime(t, "2006-01-02 15:06:05")),
					LastFailDuration: ops.Pointer(time.Minute),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetProcessingWithID(gomock.Any(), gomock.Eq(taskID)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(expTaskFromRepo, nil)

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTaskAfterTransition)).
					Do(func(ctx context.Context, _ *domain.Task) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(errors.New("test error"))

				err := d.service.Nack(ctx, taskID, nil)
				assert.EqualError(t, err, "save task: test error")
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "rollback trx error on save task error",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:05:05")

				expTaskFromRepo := &domain.Task{
					ID:          taskID,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:08:08")),
				}

				expTaskAfterTransition := &domain.Task{
					ID:               taskID,
					Status:           domain.TaskStatusFailed,
					LockedUntil:      ops.Pointer(getTime(t, "2006-01-02 15:06:05")),
					LastFailDuration: ops.Pointer(time.Minute),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetProcessingWithID(gomock.Any(), gomock.Eq(taskID)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(expTaskFromRepo, nil)

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTaskAfterTransition)).
					Do(func(ctx context.Context, _ *domain.Task) {
						assert.True(t, trx.Exists(ctx), "transaction exists")

						trx.AddRollback(ctx, func() error {
							return errors.New("test rollback trx error")
						})
					}).
					Return(errors.New("test save task error"))

				err := d.service.Nack(ctx, taskID, nil)
				assert.EqualError(t, err, "save task: test save task error")

				logs := d.logbuf.Logs()
				assert.Len(t, logs, 1)
				assert.JSONEq(t, `{
					"module": "internal/app/services/task",
					"created": "2006-01-02 15:04:05",
					"level": "error",
					"message": "rollback error on task.nack",
					"rb_error": "test rollback trx error",
					"op_error": "save task: test save task error"
				}`, logs[0])
			},
		},
		{
			name: "get processing task error",
			run: func(t *testing.T, d testDeps) {
				d.mockClock.EXPECT().
					Now().
					Return(getTime(t, "2006-01-02 15:05:05"))

				d.mockTaskRepository.EXPECT().
					GetProcessingWithID(gomock.Any(), gomock.Eq(taskID)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(nil, errors.New("test error"))

				err := d.service.Nack(ctx, taskID, nil)
				assert.EqualError(t, err, "get processing task: test error")
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "task not found",
			run: func(t *testing.T, d testDeps) {
				d.mockClock.EXPECT().
					Now().
					Return(getTime(t, "2006-01-02 15:05:05"))

				d.mockTaskRepository.EXPECT().
					GetProcessingWithID(gomock.Any(), gomock.Eq(taskID)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(nil, repository.ErrNotFound)

				err := d.service.Nack(ctx, taskID, nil)
				assert.NoError(t, err)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()

			mockClock := NewMockclock(mc)
			mockIdempotencyKeyCache := NewMockidempotencyKeyCache(mc)
			mockTaskRepository := NewMocktaskRepository(mc)
			logger, logbuf := logimpl.NewTestLogger()

			tc.run(t, testDeps{
				mockClock:               mockClock,
				mockIdempotencyKeyCache: mockIdempotencyKeyCache,
				mockTaskRepository:      mockTaskRepository,
				logbuf:                  logbuf,
				service:                 NewService(mockClock, mockIdempotencyKeyCache, mockTaskRepository, logger),
			})
		})
	}
}

func getTime(t *testing.T, value string) time.Time {
	out, err := time.Parse(time.DateTime, value)
	require.NoError(t, err)

	return out
}
