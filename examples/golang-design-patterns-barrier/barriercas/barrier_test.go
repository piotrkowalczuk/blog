package barriercas

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

//
//cases := map[string]func(int) func() func(){
//	"cas": func(i int) func() func() {
//		bar := barriercas.NewBarrier(i)
//		return func() func() {
//			return bar.Await
//		}
//	},
//		"csp": func(i int) func() func() {
//		bar := barriercsp.NewBarrier(i)
//		return func() func() {
//			return func() {
//				<-bar.Await()
//				return
//			}
//		}
//	},
//		"monitor": func(i int) func() func() {
//		bar := barriermonitor.NewBarrier(i)
//		return func() func() {
//			return bar.Await
//		}
//	},
//}
