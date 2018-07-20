package handler

import (
	"context"

	"github.com/gogolfing/cbus"
)

type updateAllStockSummary struct {
}

func NewUpdateAllStockSummary() *updateAllStockSummary {
	return &updateAllStockSummary{}
}

func (h *updateAllStockSummary) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	return nil, nil
}
