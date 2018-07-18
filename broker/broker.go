package broker

import "github.com/monetha/go-distributed/task"

// Broker is an interface for creating distributed tasks.
type Broker interface {
	// NewTask creates long-running distributed task.
	NewTask(key string, fun task.Func) (*task.Task, error)
}
