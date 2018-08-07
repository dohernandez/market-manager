package wallet

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	Finder interface {
		FindByName(name string) (*Wallet, error)
		FindByBankAccount(ba *bank.Account) (*Wallet, error)
		FindWithItemByStock(stk *stock.Stock) ([]*Wallet, error)
		LoadActiveItems(w *Wallet) error
		LoadInactiveItems(w *Wallet) error
		LoadAllItems(w *Wallet) error
		LoadItemByStock(w *Wallet, stk *stock.Stock) error
		LoadItemOperations(i *Item) error
		LoadActiveTrades(w *Wallet) error
	}

	Persister interface {
		PersistAll(ws []*Wallet) error
		PersistOperations(w *Wallet) error
		UpdateAllAccounting(ws []*Wallet) error
		UpdateAllItemsCapital(ws []*Wallet) error
	}
)
