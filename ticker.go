package duckron

import (
	"log"
	"time"
)

type Timer struct {
	ticker *time.Ticker
	stop   chan bool
}

func NewTimer(interval int) *Timer {
	timeInterval := time.Duration(interval) * time.Second
	ticker := time.NewTicker(timeInterval)
	stop := make(chan bool)

	timer := &Timer{
		ticker: ticker,
		stop:   stop,
	}

	return timer
}

func (t *Timer) Start(task func() *Error) {
	for {
		select {
		case <-t.ticker.C:
			err := task()
			if err != nil {
				log.Println("Error on timer:", err)
				return
			}
		case <-t.stop:
			log.Println("Timer killed")
			t.ticker.Stop()
			return
		}
	}
}

func (t *Timer) Stop() {
	t.stop <- true
}
