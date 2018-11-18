package barriercas

import (
	"runtime"
	"sync/atomic"
)

type Barrier struct {
	concurrency uint64
	iteration   uint64
	waiting     [2]uint64
}

func NewBarrier(i int) *Barrier {
	return &Barrier{
		concurrency: uint64(i),
		iteration:   1,
		waiting: [2]uint64{
			uint64(i),
			uint64(i),
		},
	}
}

func (b *Barrier) Await() {
	i := atomic.LoadUint64(&b.iteration)

	if atomic.AddUint64(&b.waiting[i%2], ^uint64(0)) == 0 {
		return
	}

	for {
		runtime.Gosched()

		switch atomic.LoadUint64(&b.waiting[i%2]) {
		case b.concurrency:
			// The only chance that waiting is equal to concurrency,
			// is after case bellow is already done.
			return
		case 0:
			// Last routine that reach waiting point,
			// is responsible for cleanup.
			atomic.AddUint64(&b.iteration, 1)
			atomic.StoreUint64(&b.waiting[i%2], b.concurrency)

			return
		}
	}
}
