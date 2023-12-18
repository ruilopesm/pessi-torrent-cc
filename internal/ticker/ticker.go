package ticker

import (
	"time"
)

type Ticker interface {
	Start()
	Stop()
}

type ticker struct {
	toExecute    func()
	tickInterval time.Duration
	quitChannel  chan struct{}
}

func NewTicker(tickInterval time.Duration, toExecute func()) Ticker {
	return ticker{
		toExecute:    toExecute,
		tickInterval: tickInterval,
		quitChannel:  make(chan struct{}),
	}
}

func (t ticker) Start() {
	go t.start()
}

func (t ticker) Stop() {
	t.quitChannel <- struct{}{}
	close(t.quitChannel)
}

func (t ticker) start() {
	for {
		select {
		case <-t.quitChannel:
			return
		case <-time.After(t.tickInterval):
			t.toExecute()
		}
	}
}
