package barriercsp

type Barrier struct {
	concurrency int
	iteration   chan uint8
	in          chan struct{}
	waiting     [2]chan struct{}
}

func NewBarrier(i int) *Barrier {
	iteration := make(chan uint8, i)
	for j := 0; j < i; j++ {
		iteration <- 0
	}
	return &Barrier{
		concurrency: i,
		in:          make(chan struct{}, i-1),
		waiting: [2]chan struct{}{
			make(chan struct{}),
			nil,
		},
		iteration: iteration,
	}
}

func (b *Barrier) Await() <-chan struct{} {
	i := <-b.iteration
	b.iteration <- (i + 1) % 2

	if i != 1 {
		select {
		// Write into the channel so, ...
		case b.in <- struct{}{}:
			return b.waiting[i%2]
		default:
			return b.reset(i)
		}
	}

	select {
	// ... during even iteration, it can be drained.
	case <-b.in:
		return b.waiting[i%2]
	default:
		return b.reset(i)
	}
}

func (b *Barrier) reset(i uint8) <-chan struct{} {
	b.waiting[(i+1)%2] = make(chan struct{})
	close(b.waiting[i%2])

	return b.waiting[i%2]
}
