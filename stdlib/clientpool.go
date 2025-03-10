package stdlib

import (
	"sync"
	"time"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/stdlib/debug"
)

type ClientPool[T any] struct {
	lock         sync.Mutex
	wait_channel chan T
	fn           func() T
	maxClients   int64
	curClients   int64
	waiting      int64
}

func NewClientPool[T any](maxClients int64, fn func() T) *ClientPool[T] {
	wait_channel := make(chan T, maxClients)
	return &ClientPool[T]{wait_channel: wait_channel, fn: fn, maxClients: maxClients, curClients: 0, waiting: 0}
}

func (this *ClientPool[T]) StartMetricsThread(pool_id string) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				debug.ReportMetric(pool_id+":FreeClients", len(this.wait_channel))
				debug.ReportMetric(pool_id+":CurrentWaiting", this.waiting)
			}
		}
	}()
}

func (this *ClientPool[T]) Pop() T {
	this.lock.Lock()
	if this.curClients < this.maxClients {
		defer this.lock.Unlock()
		client := this.fn()
		this.curClients += 1
		return client
	}
	this.lock.Unlock()
	this.waiting += 1
	select {
	case client := <-this.wait_channel:
		this.waiting -= 1
		return client
	}
}

func (this *ClientPool[T]) Push(client T) {
	this.wait_channel <- client
}
