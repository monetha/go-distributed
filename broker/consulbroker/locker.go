package consulbroker

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/monetha/go-distributed/locker"
)

// Locker wraps Consul distributed lock by implementing Locker interface.
type Locker struct {
	client       *api.Client
	key          string
	lockWaitTime time.Duration
	lock         *api.Lock
}

// NewLocker creates new Locker instance.
func NewLocker(client *api.Client, key string, lockWaitTime time.Duration) *Locker {
	return &Locker{
		client:       client,
		key:          key,
		lockWaitTime: lockWaitTime,
	}
}

// Key returns the name of locker.
func (l *Locker) Key() string {
	return l.key
}

// Lock attempts to acquire the locker and blocks while doing so.
// Providing a non-nil stopCh can be used to abort the locker attempt.
// Returns a channel that is closed if our locker is lost or an error.
// This channel could be closed at any time due to session invalidation,
// communication errors, operator intervention, etc. It is NOT safe to
// assume that the locker is held until Unlock(), application must be able
// to handle the locker being lost.
func (l *Locker) Lock(stopCh <-chan struct{}) (<-chan struct{}, error) {
	if l.lock != nil {
		return nil, locker.ErrLockHeld
	}

	lock, err := l.client.LockOpts(&api.LockOptions{
		Key:          l.key,
		LockWaitTime: l.lockWaitTime,
	})
	if err != nil {
		return nil, fmt.Errorf("locker: creating lock opts %s: %v", l.key, err)
	}

	lockCh, err := lock.Lock(stopCh)
	if err != nil {
		return nil, fmt.Errorf("locker: lock %s: %v", l.key, err)
	}

	if lockCh == nil {
		return nil, locker.LockCancelled(l.key)
	}

	l.lock = lock

	return lockCh, nil
}

// Unlock released the locker. It is an error to call this
// if the locker is not currently held.
func (l *Locker) Unlock() error {
	if l.lock == nil {
		return locker.ErrLockNotHeld
	}
	defer func() {
		l.lock = nil
	}()

	return l.lock.Unlock()
}
