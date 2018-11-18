package barriermonitor

import (
	"sync"
)

type Barrier struct {
	c                      *sync.Cond
	concurrency, iteration int
	waiting                [2]int
	broadcasted            [2]bool
}

func NewBarrier(n int) *Barrier {
	return &Barrier{
		c:           sync.NewCond(&sync.Mutex{}),
		concurrency: n,
	}
}

func (b *Barrier) Await() {
	b.c.L.Lock()
	i := b.iteration
	b.waiting[i%2]++
	for b.concurrency != b.waiting[i%2] {
		b.c.Wait()
	}
	if !b.broadcasted[i%2] {
		b.broadcasted[i%2] = true
		b.broadcasted[(i+1)%2] = false
		b.waiting[(i+1)%2] = 0
		b.iteration++
		b.c.Broadcast()
	}

	b.c.L.Unlock()
}
