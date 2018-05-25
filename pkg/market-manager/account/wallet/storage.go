package wallet

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	Finder interface {
		FindByName(name string) (*Wallet, error)
		FindByBankAccount(ba *bank.Account) (*Wallet, error)
		FindWalletsByStock(stk *stock.Stock) ([]*Wallet, error)
	}

	Persister interface {
		PersistAll(ws []*Wallet) error
		PersistOperations(w *Wallet) error
		UpdateAllAccounting(ws []*Wallet) error
		UpdateAccounting(w *Wallet) error
	}
)
