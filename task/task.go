package task

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/monetha/go-distributed/locker"
)

const (
	// RetryAfterErrorDelay is how long we wait before re-try after error.
	RetryAfterErrorDelay = 1 * time.Second
)

// Func is type of task function.
type Func func(context.Context) error

// Task is a long-running task (minutes to permanent).
type Task struct {
	locker    locker.Locker
	fun       Func
	wg        sync.WaitGroup
	closeOnce sync.Once
	closed    chan struct{}
}

// New creates long-running task.
func New(locker locker.Locker, fun Func) *Task {
	return (&Task{
		locker: locker,
		fun:    fun,
		closed: make(chan struct{}),
	}).runAsync()
}

func (t *Task) runAsync() *Task {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		defer log.Printf("task: task %s closed.", t.locker.Key())

		delayBeforeIteration := 0 * time.Second
		for {
			select {
			case <-t.closed:
				return
			default:
				if delayBeforeIteration > 0 {
					select {
					case <-t.closed:
						return
					case <-time.After(delayBeforeIteration):
						delayBeforeIteration = 0
					}
				}
			}

			log.Printf("task: trying to lock %s...", t.locker.Key())
			lockCh, err := t.locker.Lock(t.closed)
			if err != nil {
				log.Printf("task: lock %s: %v", t.locker.Key(), err)
				delayBeforeIteration = RetryAfterErrorDelay
				continue
			}

			log.Printf("task: locker %s acquired!", t.locker.Key())
			if err := t.runTask(lockCh); err != nil {
				log.Printf("task: locker %s: task run error: %v", t.locker.Key(), err)
				delayBeforeIteration = RetryAfterErrorDelay
			}
		}
	}()

	return t
}

func (t *Task) runTask(lockCh <-chan struct{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("task: run failed because of panic: %v: %s", r, debug.Stack())
		}
	}()

	defer func() {
		log.Printf("task: trying to unlock %s...", t.locker.Key())
		err := t.locker.Unlock()
		if err != nil {
			log.Printf("task: unlock %s: %v", t.locker.Key(), err)
		} else {
			log.Printf("task: locker %s released!", t.locker.Key())
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-t.closed:
			log.Printf("task: locker %s should be released, as task is closed.", t.locker.Key())
		case <-lockCh:
			log.Printf("task: locker %s is lost.", t.locker.Key())
		}
		cancel()
	}()

	log.Printf("task: starting task instance for locker %s.", t.locker.Key())
	defer log.Printf("task: task instance for locker %s is stopped.", t.locker.Key())
	return t.fun(ctx)
}

// Close implements io.Closer interface. It stops long running-task.
func (t *Task) Close() (err error) {
	t.closeOnce.Do(func() {
		close(t.closed)

		t.wg.Wait()
	})

	return
}
