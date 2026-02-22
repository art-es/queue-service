package consumer

import (
	"context"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/services/consumer/dto"
)

func listenTasks(ctx context.Context, tasks <-chan *domain.Task, out chan<- *dto.Message) {
	var task *domain.Task
	for {
		select {
		case <-ctx.Done():
			return
		case task = <-tasks:
		}

		out <- &dto.Message{
			Type: dto.OutputTypeTaskProcess,
			Data: dto.MessageDataTask{
				ID:        task.ID,
				Payload:   task.Payload,
				CreatedAt: task.CreatedAt,
			},
		}
	}
}
