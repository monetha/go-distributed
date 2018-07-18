package mutexbroker_test

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/monetha/go-distributed/broker/mutexbroker"
)

func ExampleNew() {
	defer log.Println("service stopped.")

	b := mutexbroker.New()

	t, err := b.NewTask("task/key", task1)
	if err != nil {
		log.Printf("New task: %v", err)
		return
	}
	defer t.Close()

	t2, err := b.NewTask("task/key", task2)
	if err != nil {
		log.Printf("New task: %v", err)
		return
	}
	defer t2.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-sigChan
}

func task1(ctx context.Context) error {
	log.Println("task1: ENTER!!")

	select {
	case <-time.After(10 * time.Second):
		log.Println("task1: EXIT!!")
		return nil
	case <-ctx.Done():
		log.Println("task1: good bye")
		return ctx.Err()
	}
}

func task2(ctx context.Context) error {
	log.Println("task2: ENTER!!")

	select {
	case <-time.After(10 * time.Second):
		log.Println("task2: EXIT!!")
		return nil
	case <-ctx.Done():
		log.Println("task2: good bye")
		return ctx.Err()
	}
}
