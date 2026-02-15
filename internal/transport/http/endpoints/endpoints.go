package endpoints

import (
	"github.com/art-es/queue-service/internal/transport/http/endpoints/v1_queues_pop"
	"github.com/art-es/queue-service/internal/transport/http/endpoints/v1_queues_push"
	"github.com/art-es/queue-service/internal/transport/http/endpoints/v1_tasks_ack"
	"github.com/art-es/queue-service/internal/transport/http/endpoints/v1_tasks_nack"
)

var (
	RegisterV1QueuesPop  = v1_queues_pop.Register
	RegisterV1QueuesPush = v1_queues_push.Register
	RegisterV1TasksAck   = v1_tasks_ack.Register
	RegisterV1TasksNack  = v1_tasks_nack.Register
)
