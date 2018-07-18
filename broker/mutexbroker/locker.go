package mutexbroker

import (
	"github.com/monetha/go-distributed/locker"
)

// Locker implements "distributed locker" using channel.
type Locker struct {
	broker   *Broker
	key      string
	leaderCh chan struct{}
}

// NewLocker creates new Locker instance.
func NewLocker(broker *Broker, key string) *Locker {
	return &Locker{
		broker: broker,
		key:    key,
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

// Unlock released the locker. It is an error to call this
// if the locker is not currently held.
func (l *Locker) Unlock() error {
	if l.leaderCh == nil {
		return locker.ErrLockNotHeld
	}

	close(l.leaderCh)
	l.leaderCh = nil

	return l.broker.unlock(l.key)
}
