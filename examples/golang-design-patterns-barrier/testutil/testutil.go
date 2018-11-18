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

func BenchmarkBarrier(b *testing.B, max int, fn func(int) interface{}) {
	for i := 4; i < max; i = i << 1 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			b.ReportAllocs()

			for n := 0; n < b.N; n++ {
				bar := fn(i)
				b.StopTimer()
				var wg sync.WaitGroup
				wg.Add(i)
				var cl func()
				switch bat := bar.(type) {
				case barrierChannel:
					cl = func() {
						<-bat.Await()
						wg.Done()
					}
				case barrier:
					cl = func() {
						bat.Await()
						wg.Done()
					}
				default:
					b.Fatal("unknown barrier interface")
				}
				b.StartTimer()

				for j := 0; j < i; j++ {
					go cl()
				}

				b.StopTimer()
				wg.Wait()
				b.StartTimer()
			}
		})
	}
}

func BenchmarkBarrierAwait(b *testing.B, max int, fn func(int) interface{}) {
	for i := 4; i < max; i = i << 1 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			b.ReportAllocs()

			for n := 0; n < b.N; n++ {
				b.StopTimer()
				bar := fn(i)
				var wg sync.WaitGroup
				wg.Add(i)
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
				// PREPARATION DONE

				for j := 0; j < i; j++ {
					go closure()
				}

				b.StopTimer()
				wg.Wait()
				b.StartTimer()
			}
		})
	}
}
