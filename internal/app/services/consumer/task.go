package consumer

import (
	"context"

	"github.com/art-es/queue-service/internal/app/domain"
	"github.com/art-es/queue-service/internal/app/domain/consumer"
)

func listenTasks(ctx context.Context, tasks <-chan *domain.Task, out chan<- *consumer.Message) {
	var task *domain.Task
	for {
		select {
		case <-ctx.Done():
			return
		case task = <-tasks:
		}

		out <- &consumer.Message{
			Type: consumer.OutputTypeTaskProcess,
			Data: consumer.MessageDataTask{
				ID:        task.ID,
				Payload:   task.Payload,
				CreatedAt: task.CreatedAt,
			},
		}
	}
}
