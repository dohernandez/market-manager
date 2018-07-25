package handler

import (
	"context"

	"github.com/gogolfing/cbus"
)

type addStocks struct {
}

func NewAddStocks() *addStocks {
	return &addStocks{}
}

func (h *addStocks) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	return nil, nil
}
