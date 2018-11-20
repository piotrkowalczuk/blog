package testutil

import (
	"fmt"
	"sync"
	"testing"
)

type barrier interface {
	Await()
}

type barrierChannel interface {
	Await() <-chan struct{}
}

func BenchmarkBarrier(b *testing.B, benchmarkConstructor bool, numberOfGoroutines, numberOfUsages int, fn func(int) interface{}) {
	for i := 4; i < numberOfGoroutines; i = i << 2 {
		for j := 1; j < numberOfUsages; j = j << 1 {
			b.Run(fmt.Sprintf("%d-%d", i, j), func(b *testing.B) {
				b.ReportAllocs()

				for n := 0; n < b.N; n++ {
					var bar interface{}
					if benchmarkConstructor {
						bar = fn(i)
						b.StopTimer()
					} else {
						b.StopTimer()
						bar = fn(i)
					}
					var wg sync.WaitGroup
					var closure func()
					switch bat := bar.(type) {
					case barrierChannel:
						closure = func() {
							<-bat.Await()
							wg.Done()
						}
					case barrier:
						closure = func() {
							bat.Await()
							wg.Done()
						}
					default:
						b.Fatal("unknown barrier interface")
					}
					b.StartTimer()

					// To enforce barrier re-usage.
					for m := 0; m < j; m++ {
						b.StopTimer()
						wg.Add(i)
						b.StartTimer()

						for g := 0; g < i; g++ {
							go closure()
						}

						b.StopTimer()
						wg.Wait()
						b.StartTimer()
					}
				}
			})
		}
	}
}

func TestBarrier(t *testing.T, fn func(i int) func() func()) {
	for size := 1; size < 101; size++ {
		for concurrency := 1; concurrency < 102; concurrency++ {
			t.Run(fmt.Sprintf("%d-%d", size, concurrency), func(t *testing.T) {
				t.Parallel()

				wait := fn(size)

				for j := 0; j < concurrency; j++ {
					in := make(chan int, size)
					for i := 0; i < cap(in); i++ {
						go func(i int) {
							in <- 1
							wait()()
							in <- 2
						}(i)
					}

					for i := 0; i < cap(in); i++ {
						if got := <-in; got != 1 {
							t.Errorf("wrong number, expected 1 but got %d", got)
						}
					}

					for i := 0; i < cap(in); i++ {
						if got := <-in; got != 2 {
							t.Errorf("wrong number, expected 1 but got %d", got)
						}
					}
				}
			})
		}
	}
}
