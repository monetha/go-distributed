package mutexbroker

import "github.com/monetha/go-distributed/locker"

var (
	_ locker.Locker = &Locker{} // ensure Locker implements Locker interface
)
