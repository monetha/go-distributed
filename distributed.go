package distributed

import (
	"github.com/monetha/go-distributed/broker"
	"github.com/monetha/go-distributed/broker/consulbroker"
	"github.com/monetha/go-distributed/broker/mutexbroker"
)

// Config is a broker configuration
type Config consulbroker.Config

// NewBroker creates a Broker which allows you to execute arbitrary tasks in a distributed infrastructure.
// The same tasks should be registered on a multiple servers, but only one instance of the task will be launched.
// Consul key is used to uniquely identify the task.
// When `cfg` is not nil Broker uses Consul to acquire the distributed lock and to ensure a single instance of
// a task is running.
// In case `cfg` is nil Broker uses mutex and runs all tasks locally allowing you not to have Consul cluster
// and simulate distributed environment.
func NewBroker(cfg *Config) (broker.Broker, error) {
	if cfg == nil {
		return mutexbroker.New(), nil
	}
	return consulbroker.New((*consulbroker.Config)(cfg))
}
