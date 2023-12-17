package ticker

import (
	"time"
)

type Ticker interface {
	Start()
	Stop()
}

type ticker struct {
	toExecute   func()
	quitChannel chan struct{}
}

func NewTicker(toExecute func()) Ticker {
	return ticker{
		toExecute:   toExecute,
		quitChannel: make(chan struct{}),
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
		case <-time.After(100 * time.Millisecond):
			t.toExecute()
		}
	}
}
