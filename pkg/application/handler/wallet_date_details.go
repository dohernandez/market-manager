package handler

import (
	"context"
	"strconv"

	"github.com/gogolfing/cbus"
	"github.com/pkg/errors"

	"time"

	"os"
	"path/filepath"

	"io"

	"github.com/satori/go.uuid"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/render"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	walletDateDetails struct {
		*walletDetails
		bankAccountFinder bank.Finder
	}
)

const getPriceConcurrency = 10

func NewWalletDateDetails(
	walletFinder wallet.Finder,
	stockFinder stock.Finder,
	dividendFinder dividend.Finder,
	ccClient *cc.Client,
	retention float64,
	bankAccountFinder bank.Finder,
) *walletDateDetails {
	return &walletDateDetails{
		walletDetails: &walletDetails{
			walletFinder:   walletFinder,
			stockFinder:    stockFinder,
			dividendFinder: dividendFinder,
			ccClient:       ccClient,
			retention:      retention,
		},
		bankAccountFinder: bankAccountFinder,
	}
}

func (h *walletDateDetails) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	walletDateDetails := command.(*appCommand.WalletDateDetails)

	wName := walletDateDetails.Wallet
	if wName == "" {
		logger.FromContext(ctx).Error("An error happen while loading wallet -> error [wallet can not be empty]")

		return nil, errors.New("missing wallet name")
	}

	date := parseOperationDateString(walletDateDetails.Date)

	w, err := h.loadWalletWithWalletItemsAndWalletTradesAtDate(
		wName,
		date,
		walletDateDetails.TransferPath,
		walletDateDetails.OperationPath,
		walletDateDetails.Excludes,
	)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	//err := h.setWalletStocksPriceAtDate(w, date)
	//if err != nil {
	//	logger.FromContext(ctx).Errorf(
	//		"An error happen while setting the price of the stocks for the date %q wallet [%s] -> error [%s]",
	//		date,
	//		wName,
	//		err,
	//	)
	//
	//	return nil, err
	//}

	var dividendsProjected []render.WalletDividendProjected

	dividendsProjected, err = h.dividendsProjectedDate(w, date)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading dividends projected -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	wDetailsOutput := h.walletDetailOutput(w, dividendsProjected)
	wDetailsOutput.WalletStockOutputs = h.walletStocksOutput(w, operation.Active)

	return wDetailsOutput, err
}
func (h *walletDateDetails) loadWalletWithWalletItemsAndWalletTradesAtDate(
	name string,
	date time.Time,
	transferPath,
	operationPath string,
	excludes []string,
) (*wallet.Wallet, error) {
	w, err := h.walletFinder.FindByName(name)
	if err != nil {
		return nil, err
	}

	err = h.walletFinder.LoadBankAccounts(w)
	if err != nil {
		return nil, err
	}

	wd := wallet.NewWallet(w.Name, w.URL)

	for _, b := range w.BankAccounts {
		wd.AddBankAccount(b)
	}

	transfers, err := h.loadTransfersUntilDate(transferPath, date)
	if err != nil {
		return nil, errors.Wrap(err, "loading transfer")
	}

	for _, t := range transfers {
		for _, b := range wd.BankAccounts {
			if t.From.ID == b.ID {
				wd.DecreaseInvestment(t.Amount)

				break
			}

			if t.To.ID == b.ID {
				wd.IncreaseInvestment(t.Amount)

				break
			}
		}
	}

	trades, ops, err := h.loadOperationUntilDate(operationPath, date)
	if err != nil {
		return nil, errors.Wrap(err, "loading operation")
	}

	for _, o := range ops {
		if o.Action == operation.Buy || o.Action == operation.Sell || o.Action == operation.Dividend {
			exclude := false
			for _, symbol := range excludes {
				if symbol == o.Stock.Symbol {
					exclude = true

					break
				}
			}

			if exclude {
				continue
			}
		}

		wd.AddOperation(o)

		nTrade, ok := trades[o.ID]
		if ok {
			n, _ := strconv.Atoi(nTrade)
			wd.AddTrade(n, o)
		} else if o.Action == operation.Dividend {
			wd.AddTrade(0, o)
		}
	}

	return wd, nil
}

func (h *walletDateDetails) loadTransfersUntilDate(importPath string, date time.Time) ([]*transfer.Transfer, error) {
	var filePaths []string

	filepath.Walk(importPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".csv" {
			filePaths = append(filePaths, path)
		}

		return nil
	})

	var transfers []*transfer.Transfer

	for _, filePath := range filePaths {
		r := util.NewCsvReader(filePath)

		r.Open()

		for {
			line, err := r.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				r.Close()

				panic(err)
			}

			t, err := createTransferFromLine(line, h.bankAccountFinder)
			if err != nil {
				r.Close()

				return nil, err
			}

			if date.After(t.Date) {
				break
			}

			transfers = append(transfers, t)
		}

		r.Close()
	}

	return transfers, nil
}

func (h *walletDateDetails) loadOperationUntilDate(importPath string, date time.Time) (map[uuid.UUID]string, []*operation.Operation, error) {
	var filePaths []string

	filepath.Walk(importPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".csv" {
			filePaths = append(filePaths, path)
		}

		return nil
	})

	var operations []*operation.Operation
	trades := map[uuid.UUID]string{}

	for _, filePath := range filePaths {
		r := util.NewCsvReader(filePath)

		r.Open()

		for {
			line, err := r.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				r.Close()

				panic(err)
			}

			o, err := createOperationFromLine(line, h.stockFinder)

			if err != nil {
				r.Close()

				return nil, nil, err
			}

			//fmt.Printf("operation [%s] date.After(o.Date) date [%+v] after o.Date [%+v]\n", o.ID, date, o.Date)
			if o.Date.After(date) {
				break
			}
			//fmt.Printf("operation %+v\n", o)

			operations = append(operations, o)

			if line[0] != "" {
				trades[o.ID] = line[0]
			}
		}

		r.Close()
	}

	return trades, operations, nil
}

func (h *walletDateDetails) setWalletStocksPriceAtDate(w *wallet.Wallet, date time.Time) error {
	//var wg sync.WaitGroup
	//
	//for key, value := range w {
	//
	//}

	return nil
}
