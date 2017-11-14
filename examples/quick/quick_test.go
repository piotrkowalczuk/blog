package main

import (
	"bytes"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"testing/quick"
)

type request struct {
	*http.Request
}

// Generate implements quick Generator interface.
func (r *request) Generate(rand *rand.Rand, size int) reflect.Value {
	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}
	req, _ := http.NewRequest(methods[rand.Intn(9)], "http://example.com", bytes.NewBufferString(strconv.FormatInt(rand.Int63n(9), 10)))
	return reflect.ValueOf(&request{Request: req})
}

func TestSomething(t *testing.T) {
	fn1 := func(r *request) bool {
		err := func(r *request) error {
			return
		}(r)
		return err == nil
	}

	if err := quick.Check(fn1, &quick.Config{
		MaxCount: 100,
	}); err != nil {
		t.Fatal(err)
	}
}
