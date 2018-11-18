package barrierchannel

import (
	"testing"

	"github.com/piotrkowalczuk/blog/examples/golang-design-patterns-barrier/testutil"
)

var max = 4096 + 1

func BenchmarkBarrier(b *testing.B) {
	testutil.BenchmarkBarrier(b, max, func(i int) interface{} {
		return NewBarrier(i)
	})
}

func BenchmarkBarrier_Await(b *testing.B) {
	testutil.BenchmarkBarrierAwait(b, max, func(i int) interface{} {
		return NewBarrier(i)
	})
}
