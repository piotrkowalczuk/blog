package barriermonitor

import (
	"testing"

	"github.com/piotrkowalczuk/blog/examples/golang-design-patterns-barrier/testutil"
)

var max = 4096 + 1

func BenchmarkBarrier(b *testing.B) {
	testutil.BenchmarkBarrier(b, true, max, 32, func(i int) interface{} {
		return NewBarrier(i)
	})
}

func BenchmarkBarrier_Await(b *testing.B) {
	testutil.BenchmarkBarrier(b, false, max, 32, func(i int) interface{} {
		return NewBarrier(i)
	})
}

func TestBarrier(t *testing.T) {
	testutil.TestBarrier(t, func(i int) func() func() {
		bar := NewBarrier(i)
		return func() func() {
			return bar.Await
		}
	})
}
