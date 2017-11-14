package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type example struct {
	A int64
	B string
	C float64
	D []byte
	E []string
	F []int64
	G []float64
	H map[string]string
	I map[int64]int64
}

var prepare = func(w io.Writer) {
	var arr []*example
	for i := 0; i < 200; i++ {
		arr = append(arr, &example{
			A: 10,
			B: "example",
			C: 100.100,
			D: []byte("example"),
			E: []string{"a", "b", "c", "d"},
			F: []int64{1, 2, 3, 4},
			G: []float64{100.100, 200.200, 300.300, 400.400},
			H: map[string]string{
				"A": "A",
				"B": "B",
				"C": "C",
			},
			I: map[int64]int64{
				1: 1,
				2: 2,
				3: 3,
			},
		})
	}
	err := json.NewEncoder(w).Encode(arr)
	if err != nil {
		panic(err)
	}
	return
}

var benchExample []*example

func BenchmarkMarshall(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prepare(rw)
	}))
	for n := 0; n < b.N; n++ {

		func() {
			res := response(b, s)
			buf, err := ioutil.ReadAll(res.Body)
			if err != nil {
				b.Fatalf("unexpected error: %s", err.Error())
			}
			var obj []*example
			if err = json.Unmarshal(buf, &obj); err != nil {
				b.Fatalf("unexpected error: %s", err.Error())
			}
			benchExample = obj
		}()
	}
}

func BenchmarkDecode(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prepare(rw)
	}))
	for n := 0; n < b.N; n++ {
		func() {
			res := response(b, s)
			var obj []*example
			if err := json.NewDecoder(res.Body).Decode(&obj); err != nil {
				b.Fatalf("unexpected error: %s", err.Error())
			}
			if len(obj) == 0 {
				b.Fatal("to short")
			}
			benchExample = obj
		}()
	}
}

func response(b *testing.B, s *httptest.Server) *http.Response {
	b.StopTimer()
	req, err := http.NewRequest(http.MethodGet, s.URL, nil)
	if err != nil {
		b.Fatal(err)
	}
	res, err := s.Client().Do(req)
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	return res
}
