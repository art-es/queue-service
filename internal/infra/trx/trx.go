package trx

import (
	"errors"
	"sync"
)

func newTrx() *trx {
	return &trx{
		values:        make(map[any]any),
		rollbackFuncs: []func() error{},
		commitFuncs:   []func() error{},
		mu:            sync.RWMutex{},
	}
}

type trx struct {
	values        map[any]any
	rollbackFuncs []func() error
	commitFuncs   []func() error
	mu            sync.RWMutex
}

func (t *trx) value(key any) (any, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	v, ok := t.values[key]
	return v, ok
}

func (t *trx) setValue(key, value any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.values[key] = value
}

func (t *trx) addRollback(fn func() error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rollbackFuncs = append(t.rollbackFuncs, fn)
}

func (t *trx) addCommit(fn func() error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.commitFuncs = append(t.commitFuncs, fn)
}

func (t *trx) rollback() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, fn := range t.rollbackFuncs {
		err = errors.Join(err, fn())
	}
	return
}

func (t *trx) commit() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, fn := range t.commitFuncs {
		err = errors.Join(err, fn())
	}
	return
}
