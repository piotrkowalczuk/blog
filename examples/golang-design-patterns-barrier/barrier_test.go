package barrier_test

import (
	"testing"

	"github.com/piotrkowalczuk/blog/examples/golang-design-patterns-barrier/barriercas"
	"github.com/piotrkowalczuk/blog/examples/golang-design-patterns-barrier/barriercsp"
	"github.com/piotrkowalczuk/blog/examples/golang-design-patterns-barrier/barriermonitor"
)

func TestBarrier(t *testing.T) {
	cases := map[string]func(int) func() func(){
		"cas": func(i int) func() func() {
			bar := barriercas.NewBarrier(i)
			return func() func() {
				return bar.Await
			}
		},
		"csp": func(i int) func() func() {
			bar := barriercsp.NewBarrier(i)
			return func() func() {
				return func() {
					<-bar.Await()
					return
				}
			}
		},
		"monitor": func(i int) func() func() {
			bar := barriermonitor.NewBarrier(i)
			return func() func() {
				return bar.Await
			}
		},
	}

	for size := 1; size < 101; size++ {
		for concurrency := 1; concurrency < 102; concurrency++ {
			for hint, fn := range cases {
				t.Run(hint, func(t *testing.T) {
					//t.Parallel()

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
}
