package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/repository"
	"github.com/art-es/queue-service/internal/infra/log"
	"github.com/art-es/queue-service/internal/infra/log/logimpl"
	"github.com/art-es/queue-service/internal/infra/ops"
	"github.com/art-es/queue-service/internal/infra/trx"
)

func TestService_Push(t *testing.T) {
	var (
		ctx            = context.Background()
		taskID         = "testTaskID"
		queueName      = "testQueueName"
		payload        = "testPayload"
		idempotencyKey = "testIdempotencyKey"

		mockTaskSave = func(_ context.Context, task *domain.Task) {
			task.ID = taskID
		}
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
			name: "push new task",
			run: func(t *testing.T, d testDeps) {
				expTaskBeforeSave := &domain.Task{
					QueueName: queueName,
					Payload:   payload,
					Status:    domain.TaskStatusPending,
				}

				expTaskAfterSave := &domain.Task{
					ID:        taskID,
					QueueName: queueName,
					Payload:   payload,
					Status:    domain.TaskStatusPending,
				}

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTaskBeforeSave)).
					Do(mockTaskSave).
					Return(nil)

				task, err := d.service.Push(ctx, &PushRequest{
					QueueName: queueName,
					Payload:   payload,
				})

				assert.NoError(t, err)
				assert.Equal(t, expTaskAfterSave, task)
			},
		},
		{
			name: "push new task with idempotency key",
			run: func(t *testing.T, d testDeps) {
				expTaskBeforeSave := &domain.Task{
					QueueName: queueName,
					Payload:   payload,
					Status:    domain.TaskStatusPending,
				}

				expTaskAfterSave := &domain.Task{
					ID:        taskID,
					QueueName: queueName,
					Payload:   payload,
					Status:    domain.TaskStatusPending,
				}

				d.mockIdempotencyKeyCache.EXPECT().
					GetQueuePush(gomock.Eq(idempotencyKey)).
					Return(nil, false)

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTaskBeforeSave)).
					Do(mockTaskSave).
					Return(nil)

				d.mockIdempotencyKeyCache.EXPECT().
					SetQueuePush(gomock.Eq(idempotencyKey), gomock.Eq(expTaskAfterSave))

				task, err := d.service.Push(ctx, &PushRequest{
					QueueName:      queueName,
					Payload:        payload,
					IdempotencyKey: ops.Pointer(idempotencyKey),
				})

				assert.NoError(t, err)
				assert.Equal(t, expTaskAfterSave, task)
			},
		},
		{
			name: "idempotency key exists in cache",
			run: func(t *testing.T, d testDeps) {
				cachedTask := &domain.Task{
					ID:        taskID,
					QueueName: queueName,
					Payload:   payload,
					Status:    domain.TaskStatusPending,
				}

				d.mockIdempotencyKeyCache.EXPECT().
					GetQueuePush(gomock.Eq(idempotencyKey)).
					Return(cachedTask, true)

				task, err := d.service.Push(ctx, &PushRequest{
					QueueName:      queueName,
					Payload:        payload,
					IdempotencyKey: ops.Pointer(idempotencyKey),
				})

				assert.NoError(t, err)
				assert.Equal(t, cachedTask, task)
			},
		},
		{
			name: "task save error",
			run: func(t *testing.T, d testDeps) {
				expTask := &domain.Task{
					QueueName: queueName,
					Payload:   payload,
					Status:    domain.TaskStatusPending,
				}

				d.mockTaskRepository.EXPECT().
					Save(gomock.Any(), gomock.Eq(expTask)).
					Do(mockTaskSave).
					Return(errors.New("test error"))

				task, err := d.service.Push(ctx, &PushRequest{
					QueueName: queueName,
					Payload:   payload,
				})

				assert.EqualError(t, err, "save task: test error")
				assert.Nil(t, task)
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

func TestService_Pop(t *testing.T) {
	var (
		ctx       = context.Background()
		taskID    = "testTaskID"
		queueName = "testQueueName"
	)

	type testDeps struct {
		mockClock          *Mockclock
		mockTaskRepository *MocktaskRepository
		logbuf             log.Buffer
		service            *Service
	}

	for _, tc := range []struct {
		name string
		run  func(t *testing.T, d testDeps)
	}{
		{
			name: "pop task",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:04:05")

				expTaskFromRepo := &domain.Task{
					ID:        taskID,
					QueueName: queueName,
					Status:    domain.TaskStatusPending,
				}

				expTaskAfterTransition := &domain.Task{
					ID:          taskID,
					QueueName:   queueName,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:09:05")),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetFirstPending(gomock.Any(), gomock.Eq(queueName)).
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

				task, err := d.service.Pop(ctx, queueName)

				assert.NoError(t, err)
				assert.Equal(t, expTaskAfterTransition, task)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "commit trx error",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:04:05")

				expTaskFromRepo := &domain.Task{
					ID:        taskID,
					QueueName: queueName,
					Status:    domain.TaskStatusPending,
				}

				expTaskAfterTransition := &domain.Task{
					ID:          taskID,
					QueueName:   queueName,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:09:05")),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetFirstPending(gomock.Any(), gomock.Eq(queueName)).
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

				task, err := d.service.Pop(ctx, queueName)

				assert.EqualError(t, err, "commit trx: test error")
				assert.Nil(t, task)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "save task error",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:04:05")

				expTaskFromRepo := &domain.Task{
					ID:        taskID,
					QueueName: queueName,
					Status:    domain.TaskStatusPending,
				}

				expTaskAfterTransition := &domain.Task{
					ID:          taskID,
					QueueName:   queueName,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:09:05")),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetFirstPending(gomock.Any(), gomock.Eq(queueName)).
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

				task, err := d.service.Pop(ctx, queueName)

				assert.EqualError(t, err, "save task: test error")
				assert.Nil(t, task)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "rollback trx error on save task error",
			run: func(t *testing.T, d testDeps) {
				now := getTime(t, "2006-01-02 15:04:05")

				expTaskFromRepo := &domain.Task{
					ID:        taskID,
					QueueName: queueName,
					Status:    domain.TaskStatusPending,
				}

				expTaskAfterTransition := &domain.Task{
					ID:          taskID,
					QueueName:   queueName,
					Status:      domain.TaskStatusProcessing,
					LockedUntil: ops.Pointer(getTime(t, "2006-01-02 15:09:05")),
				}

				d.mockClock.EXPECT().
					Now().
					Return(now)

				d.mockTaskRepository.EXPECT().
					GetFirstPending(gomock.Any(), gomock.Eq(queueName)).
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

				task, err := d.service.Pop(ctx, queueName)

				assert.EqualError(t, err, "save task: test save task error")
				assert.Nil(t, task)

				logs := d.logbuf.Logs()
				assert.Len(t, logs, 1)
				assert.JSONEq(t, `{
					"module": "internal/app/services/queue",
					"created": "2006-01-02 15:04:05",
					"level": "error",
					"message": "rollback error on queue.pop",
					"rb_error": "test rollback trx error",
					"op_error": "save task: test save task error"
				}`, logs[0])
			},
		},
		{
			name: "get first pending task error",
			run: func(t *testing.T, d testDeps) {
				d.mockClock.EXPECT().
					Now().
					Return(getTime(t, "2006-01-02 15:04:05"))

				d.mockTaskRepository.EXPECT().
					GetFirstPending(gomock.Any(), gomock.Eq(queueName)).
					Do(func(ctx context.Context, _ string) {
						assert.True(t, trx.Exists(ctx), "transaction exists")
					}).
					Return(nil, errors.New("test error"))

				task, err := d.service.Pop(ctx, queueName)

				assert.EqualError(t, err, "get first pending task: test error")
				assert.Nil(t, task)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
		{
			name: "no pending task",
			run: func(t *testing.T, d testDeps) {
				d.mockClock.EXPECT().
					Now().
					Return(getTime(t, "2006-01-02 15:04:05"))

				d.mockTaskRepository.EXPECT().
					GetFirstPending(gomock.Any(), gomock.Eq(queueName)).
					Return(nil, repository.ErrNotFound)

				task, err := d.service.Pop(ctx, queueName)

				assert.NoError(t, err)
				assert.Nil(t, task)
				assert.Empty(t, d.logbuf.Logs())
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()

			mockClock := NewMockclock(mc)
			mockTaskRepository := NewMocktaskRepository(mc)
			logger, logbuf := logimpl.NewTestLogger()

			tc.run(t, testDeps{
				mockClock:          mockClock,
				mockTaskRepository: mockTaskRepository,
				logbuf:             logbuf,
				service:            NewService(mockClock, nil, mockTaskRepository, logger),
			})
		})
	}
}

func getTime(t *testing.T, value string) time.Time {
	out, err := time.Parse(time.DateTime, value)
	require.NoError(t, err)

	return out
}
