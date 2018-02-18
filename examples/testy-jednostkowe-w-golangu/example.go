package example

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

//go:generate mockery -case=underscore -all

type PutCarPayload struct {
	ID           int64 `storage:"identifier"` // if set, storage will perform update, otherwise insert
	Name         string
	Age, Mileage int
	Owner        string
}

type PutCarController struct {
	Storage Storage
}

func (pcc *PutCarController) Handle(req *http.Request) (interface{}, error) {
	var pay PutCarPayload
	if err := json.NewDecoder(req.Body).Decode(&pay); err != nil {
		return nil, errors.New("invalid json payload")
	}
	req.Body.Close()

	if pay.Name == "" {
		return nil, errors.New("missing name")
	}

	if err := pcc.Storage.Put(req.Context(), &pay); err != nil {
		return nil, err
	}

	return &pay, nil
}

type Storage interface {
	Put(context.Context, interface{}) error
	Get(context.Context, int64) (interface{}, error)
}
