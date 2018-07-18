package mutexbroker

import (
	"fmt"
	"sync"

	"github.com/monetha/go-distributed/locker"
	"github.com/monetha/go-distributed/task"
)

// Broker can start long-running tasks in locally.
type Broker struct {
	mutexMap   map[string]chan struct{}
	localMutex sync.Mutex
}

// New creates new broker that holds named mutexes which are used to start long-running task.
func New() *Broker {
	return &Broker{
		mutexMap: make(map[string]chan struct{}),
	}
}

// NewTask creates new long-running task in current process.
func (b *Broker) NewTask(key string, fun task.Func) (*task.Task, error) {
	return task.New(newMutexLocker(b, key), fun), nil
}

func (b *Broker) lock(name string, stopCh <-chan struct{}) error {
	b.localMutex.Lock()
	mc, ok := b.mutexMap[name]
	if !ok {
		mc = make(chan struct{}, 1)
		b.mutexMap[name] = mc
	}
	b.localMutex.Unlock()

	select {
	case <-stopCh:
		return locker.LockCancelled(name)
	case mc <- struct{}{}:
		return nil
	}
}

func (b *Broker) unlock(name string) error {
	b.localMutex.Lock()
	defer b.localMutex.Unlock()
	mc := b.mutexMap[name]
	if len(mc) == 0 {
		return fmt.Errorf("locker: no named mutex acquired: %v", name)
	}
	<-mc
	return nil
}

func newMutexLocker(broker *Broker, key string) *mutexLocker {
	return &mutexLocker{
		broker: broker,
		key:    key,
	}
}

type mutexLocker struct {
	broker   *Broker
	key      string
	leaderCh chan struct{}
}

func (l *mutexLocker) Key() string {
	return l.key
}

func (l *mutexLocker) Lock(stopCh <-chan struct{}) (<-chan struct{}, error) {
	if l.leaderCh != nil {
		return nil, locker.ErrLockHeld
	}

	err := l.broker.lock(l.key, stopCh)
	if err != nil {
		return nil, err
	}

	l.leaderCh = make(chan struct{})
	return l.leaderCh, nil
}

func (l *mutexLocker) Unlock() error {
	if l.leaderCh == nil {
		return locker.ErrLockNotHeld
	}

	close(l.leaderCh)
	l.leaderCh = nil

	return l.broker.unlock(l.key)
}
