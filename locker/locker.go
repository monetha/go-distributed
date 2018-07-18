package locker

import "fmt"

var (
	// ErrLockHeld is returned if we attempt to double locker
	ErrLockHeld = fmt.Errorf("lock already held")

	// ErrLockNotHeld is returned if we attempt to unlock a locker
	// that we do not hold.
	ErrLockNotHeld = fmt.Errorf("lock not held")

	// ErrLockCancelled is returned when attempt to lock is cancelled.
	ErrLockCancelled = fmt.Errorf("lock cancelled")
)

// Locker is an interface of distributed locker.
type Locker interface {
	// Key returns the name of locker.
	Key() string

	// Lock attempts to acquire the locker and blocks while doing so.
	// Providing a non-nil stopCh can be used to abort the locker attempt.
	// Returns a channel that is closed if our locker is lost or an error.
	// This channel could be closed at any time due to session invalidation,
	// communication errors, operator intervention, etc. It is NOT safe to
	// assume that the locker is held until Unlock(), application must be able
	// to handle the locker being lost.
	Lock(stopCh <-chan struct{}) (<-chan struct{}, error)

	// Unlock released the locker. It is an error to call this
	// if the locker is not currently held.
	Unlock() error
}

// LockCancelled returns error for cancelled lock.
func LockCancelled(name string) error {
	return fmt.Errorf("locker: lock %s: cancelled", name)
}
