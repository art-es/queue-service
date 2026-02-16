package inmemory

import (
	"sync"
	"time"

	"github.com/art-es/queue-service/internal/app/domain"
)

const cacheTTL = time.Minute

type IdempotencyKeyCache struct {
	mapQueuePush sync.Map
	mapTaskAck   sync.Map
	mapTaskNack  sync.Map
}

func NewIdempotencyKeyCache() *IdempotencyKeyCache {
	return &IdempotencyKeyCache{
		mapQueuePush: sync.Map{},
		mapTaskAck:   sync.Map{},
		mapTaskNack:  sync.Map{},
	}
}

func (c *IdempotencyKeyCache) GetQueuePush(key string) (*domain.Task, bool) {
	if v, ok := c.mapQueuePush.Load(key); ok {
		return v.(*domain.Task), true
	}
	return nil, false
}

func (c *IdempotencyKeyCache) SetQueuePush(key string, result *domain.Task) {
	c.mapQueuePush.Store(key, result)

	go func() {
		<-time.After(cacheTTL)
		c.mapQueuePush.Delete(key)
	}()
}

func (c *IdempotencyKeyCache) HasTaskAck(key string) bool {
	_, ok := c.mapTaskAck.Load(key)
	return ok
}

func (c *IdempotencyKeyCache) SetTaskAck(key string) {
	c.mapTaskAck.Store(key, struct{}{})

	go func() {
		<-time.After(cacheTTL)
		c.mapTaskAck.Delete(key)
	}()
}

func (c *IdempotencyKeyCache) HasTaskNack(key string) bool {
	_, ok := c.mapTaskNack.Load(key)
	return ok
}

func (c *IdempotencyKeyCache) SetTaskNack(key string) {
	c.mapTaskNack.Store(key, struct{}{})

	go func() {
		<-time.After(cacheTTL)
		c.mapTaskNack.Delete(key)
	}()
}
