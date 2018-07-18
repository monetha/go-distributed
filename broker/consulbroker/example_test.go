package consulbroker_test

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/monetha/go-distributed/broker/consulbroker"
)

func ExampleNew() {
	defer log.Println("service stopped.")

	b, err := consulbroker.New(&consulbroker.Config{Address: "127.0.0.1:8500"})
	if err != nil {
		log.Printf("New broker: %v", err)
		return
	}

	t, err := b.NewTask("some/long/running/task", someLongRunningTask)
	if err != nil {
		log.Printf("New task: %v", err)
		return
	}
	defer t.Close()

	t2, err := b.NewTask("other/long/running/task", otherLongRunningTask)
	if err != nil {
		log.Printf("New task: %v", err)
		return
	}
	defer t2.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigChan
}

func someLongRunningTask(ctx context.Context) error {
	log.Println("someLongRunningTask: I'm alive!!")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("someLongRunningTask: Doing some interesting stuff..")
		case <-ctx.Done():
			log.Println("someLongRunningTask: Oh, no! I should stop my work.")
			return ctx.Err()
		}
	}
}

func otherLongRunningTask(ctx context.Context) error {
	log.Println("otherLongRunningTask: BANG BANG!!")

	select {
	case <-time.After(10 * time.Second):
		log.Println("otherLongRunningTask: BOOO!!")
		return nil
	case <-ctx.Done():
		log.Println("otherLongRunningTask: good bye")
		return ctx.Err()
	}
}
