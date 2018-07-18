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
	return task.New(NewLocker(b, key), fun), nil
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
