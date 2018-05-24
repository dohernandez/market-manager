package transfer

import (
	"github.com/satori/go.uuid"

	"time"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
)

type (
	Transfer struct {
		ID     uuid.UUID
		From   *bank.Account
		To     *bank.Account
		Amount mm.Value
		Date   time.Time
	}
)

func NewTransfer(From, To *bank.Account, amount float64, date time.Time) *Transfer {
	return &Transfer{
		ID:     uuid.NewV4(),
		From:   From,
		To:     To,
		Amount: mm.Value{Amount: amount},
		Date:   date,
	}
}
