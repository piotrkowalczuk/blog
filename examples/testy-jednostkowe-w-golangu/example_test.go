package example_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/piotrkowalczuk/blog/examples/testy-jednostkowe-w-golangu"
	"github.com/piotrkowalczuk/blog/examples/testy-jednostkowe-w-golangu/mocks"
	"github.com/stretchr/testify/mock"
)

func TestPutCarController_Handle(t *testing.T) {
	storage := &mocks.Storage{}

	req := func(t *testing.T, obj interface{}) *http.Request {
		buf, err := json.Marshal(obj)
		if err != nil {
			t.Fatalf("payload marshal failure: %s", err.Error())
		}
		return httptest.NewRequest(http.MethodPut, "http://localhost", bytes.NewReader(buf))
	}
	cases := map[string]struct {
		req  *http.Request
		init func(*testing.T)
		res  interface{}
		err  error
	}{
		"missing-name": {
			req: req(t, &example.PutCarPayload{}),
			err: errors.New("missing name"),
		},
		"success": {
			req: req(t, &example.PutCarPayload{
				Name: "brand new car",
			}),
			res: &example.PutCarPayload{
				ID:   100,
				Name: "brand new car",
			},
			init: func(t *testing.T) {
				storage.On("Put", mock.Anything, mock.AnythingOfType("*example.PutCarPayload")).
					Run(func(args mock.Arguments) {
						if pay, ok := args.Get(1).(*example.PutCarPayload); ok {
							pay.ID = 100
						}
					}).
					Return(nil).
					Once()
			},
		},
		"deadline-exceeded": {
			req: req(t, &example.PutCarPayload{
				Name: "brand new car",
			}),
			err: context.DeadlineExceeded,
			init: func(t *testing.T) {
				storage.On("Put", mock.Anything, mock.AnythingOfType("*example.PutCarPayload")).
					Return(context.DeadlineExceeded).
					Once()
			},
		},
		"text-payload": {
			req: httptest.NewRequest(http.MethodPut, "http://localhost", strings.NewReader("not-json-at-all")),
			err: errors.New("invalid json payload"),
		},
	}

	for hint, given := range cases {
		t.Run(hint, func(t *testing.T) {
			storage.ExpectedCalls = nil

			if given.init != nil {
				given.init(t)
			}

			got, err := (&example.PutCarController{
				Storage: storage,
			}).Handle(given.req)
			if given.err != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if given.err.Error() != err.Error() {
					t.Fatalf("wrong error, expected:\n	%s\nbut got:\n	%s", given.err.Error(), err.Error())
				}
			} else {
				if !reflect.DeepEqual(given.res, got) {
					t.Fatalf("wrong response, expected:\n	%v\nbut got:\n	%v", given.res, got)
				}
			}

			mock.AssertExpectationsForObjects(t, storage)
		})
	}
}
