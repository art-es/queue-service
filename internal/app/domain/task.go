package domain

import (
	"time"

	"github.com/art-es/queue-service/internal/infra/ops"
)

const (
	TaskStatusPending    = "pending"    // Ready for action
	TaskStatusProcessing = "processing" // Taken into action
	TaskStatusFailed     = "failed"     // Action failed
)

const (
	taskProcessingTimeout = 5 * time.Minute
	taskFirstFailTimeout  = 1 * time.Minute
)

type Task struct {
	ID               string
	QueueName        string
	Payload          string
	Status           string
	CreatedAt        time.Time
	LockedUntil      *time.Time
	LastFailDuration *time.Duration
}

func NewTask(queueName, payload string) *Task {
	return &Task{
		QueueName: queueName,
		Payload:   payload,
		Status:    TaskStatusPending,
	}
}

func (t *Task) ToProcessing(now time.Time) {
	t.Status = TaskStatusPending
	t.LockedUntil = ops.Pointer(now.Add(taskProcessingTimeout))
}

func (t *Task) ToFailed(now time.Time) {
	lockDuration := taskFirstFailTimeout
	if t.LastFailDuration != nil {
		lockDuration = *t.LastFailDuration
	}

	t.Status = TaskStatusFailed
	t.LockedUntil = ops.Pointer(now.Add(lockDuration))
	t.LastFailDuration = ops.Pointer(lockDuration)
}
