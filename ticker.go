package duckron

import (
	"log"
	"time"
)

type Timer struct {
	ticker *time.Ticker
	stop   chan bool
}

func NewTimer(interval time.Duration) *Timer {
	ticker := time.NewTicker(interval)
	stop := make(chan bool)

	timer := &Timer{
		ticker: ticker,
		stop:   stop,
	}

	return timer
}

func (t *Timer) Start(task func() *Error) *Error {
	for {
		select {
		case <-t.ticker.C:
			err := task()
			if err != nil {
				log.Println("Error on timer:", err)
				return err
			}
		case <-t.stop:
			log.Println("Timer killed")
			t.ticker.Stop()
			return nil
		}
	}
}

func (t *Timer) Stop() {
	t.stop <- true
}
