package domain

import (
	"testing"
	"time"

	"github.com/art-es/queue-service/internal/infra/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	queueName := "testQueueName"
	payload := "testPayload"

	task := NewTask(queueName, payload)

	expTask := &Task{
		QueueName: queueName,
		Payload:   payload,
		Status:    TaskStatusPending,
	}

	assert.Equal(t, expTask, task)
}

func TestTask_ToProcessing(t *testing.T) {
	now, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
	require.NoError(t, err)

	task := &Task{
		Status:      TaskStatusPending,
		LockedUntil: nil,
	}

	task.ToProcessing(now)

	expLockedUntil, err := time.Parse(time.DateTime, "2006-01-02 15:09:05")
	require.NoError(t, err)

	expTask := &Task{
		Status:      TaskStatusProcessing,
		LockedUntil: &expLockedUntil,
	}

	assert.Equal(t, expTask, task)
}

func TestTask_ToFailed(t *testing.T) {
	now, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
	require.NoError(t, err)

	t.Run("LastFailDuration is nil", func(t *testing.T) {
		task := &Task{
			Status:      TaskStatusPending,
			LockedUntil: nil,
		}

		task.ToFailed(now)

		expLockedUntil, err := time.Parse(time.DateTime, "2006-01-02 15:05:05")
		require.NoError(t, err)

		expTask := &Task{
			Status:           TaskStatusFailed,
			LockedUntil:      &expLockedUntil,
			LastFailDuration: ops.Pointer(taskFirstFailTimeout),
		}

		assert.Equal(t, expTask, task)
	})

	t.Run("LastFailDuration is not nil", func(t *testing.T) {
		task := &Task{
			Status:           TaskStatusPending,
			LockedUntil:      nil,
			LastFailDuration: ops.Pointer(3 * time.Minute),
		}

		task.ToFailed(now)

		expLockedUntil, err := time.Parse(time.DateTime, "2006-01-02 15:10:05")
		require.NoError(t, err)

		expTask := &Task{
			Status:           TaskStatusFailed,
			LockedUntil:      &expLockedUntil,
			LastFailDuration: ops.Pointer(6 * time.Minute),
		}

		assert.Equal(t, expTask, task)
	})
}
