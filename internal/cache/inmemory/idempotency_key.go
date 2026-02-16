package inmemory

import "github.com/art-es/queue-service/internal/app/domain"

// TODO: make concurrent safety
type IdempotencyKeyCache struct {
	mapQueuePush map[string]*domain.Task
	mapTaskAck   map[string]struct{}
	mapTaskNack  map[string]struct{}
}

func NewIdempotencyKeyCache() *IdempotencyKeyCache {
	return &IdempotencyKeyCache{
		mapQueuePush: map[string]*domain.Task{},
		mapTaskAck:   map[string]struct{}{},
		mapTaskNack:  map[string]struct{}{},
	}
}

func (c *IdempotencyKeyCache) GetQueuePush(key string) (*domain.Task, bool) {
	v, ok := c.mapQueuePush[key]
	return v, ok
}

func (c *IdempotencyKeyCache) SetQueuePush(key string, result *domain.Task) {
	c.mapQueuePush[key] = result
}

func (c *IdempotencyKeyCache) HasTaskAck(key string) bool {
	_, ok := c.mapTaskAck[key]
	return ok
}

func (c *IdempotencyKeyCache) SetTaskAck(key string) {
	c.mapTaskAck[key] = struct{}{}
}

func (c *IdempotencyKeyCache) HasTaskNack(key string) bool {
	_, ok := c.mapTaskNack[key]
	return ok
}

func (c *IdempotencyKeyCache) SetTaskNack(key string) {
	c.mapTaskNack[key] = struct{}{}
}
