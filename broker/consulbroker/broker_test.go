package consulbroker

import "github.com/monetha/go-distributed/broker"

var (
	_ broker.Broker = &Broker{} // ensure Broker implements Broker interface
)
