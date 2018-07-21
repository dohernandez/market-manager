package handler

import (
	"context"

	"github.com/gogolfing/cbus"
)

type addStocks struct {
}

func NewAddAllStocks() *addStocks {
	return &addStocks{}
}

func (h *addStocks) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
}
