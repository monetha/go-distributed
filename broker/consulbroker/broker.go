package consulbroker

import (
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/monetha/go-distributed/task"
)

const (
	// LockWaitTime is how long we block for at a time to check if locker
	// acquisition is possible. This affects the minimum time it takes to cancel
	// a Lock acquisition.
	LockWaitTime = 2 * time.Second
)

// Config is used to configure the creation of a broker
type Config struct {
	// Address is the address of the Consul server
	Address string

	// Scheme is the URI scheme for the Consul server
	Scheme string

	// Token is used to provide a per-request ACL token
	Token string
}

// Broker can start long-running tasks (minutes to permanent) in Consul cluster.
type Broker struct {
	client *api.Client
}

// New creates a broker that can start long-running tasks in Consul cluster.
func New(config *Config) (*Broker, error) {
	if config == nil {
		config = &Config{}
	}

	client, err := api.NewClient(&api.Config{
		Address: config.Address,
		Scheme:  config.Scheme,
		Token:   config.Token,
	})
	if err != nil {
		return nil, err
	}

	return &Broker{client: client}, nil
}

// NewTask creates new long-running task in Consul cluster.
// Task makes a best effort to ensure that exactly one instance of a task is executing in a cluster.
// Task may be re-started when needed until it's been closed.
func (b *Broker) NewTask(key string, fun task.Func) (*task.Task, error) {
	return task.New(NewLocker(b.client, key, LockWaitTime), fun), nil
}
