package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

func createTransferFromLine(line []string, bankAccountFinder bank.Finder) (*transfer.Transfer, error) {

	date := parseTransferDateString(line[0])

	from, err := bankAccountFinder.FindByAlias(line[1])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %q", err.Error(), line[1]))
	}

	to, err := bankAccountFinder.FindByAlias(line[2])
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s %q", err.Error(), line[2]))
	}

	amount, err := parseTransferPriceString(line[3])
	if err != nil {
		return nil, err
	}

	t := transfer.NewTransfer(from, to, amount, date)

	return t, nil
}

// parseTransferDateString - parse a potentially partial date string to Time
func parseTransferDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parseTransferPriceString - parse a potentially float string to float64
func parseTransferPriceString(price string) (float64, error) {
	price = strings.Replace(price, ".", "", 1)
	price = strings.Replace(price, ",", ".", 1)

	return strconv.ParseFloat(price, 64)
}

func createOperationFromLine(line []string, stockFinder stock.Finder) (*operation.Operation, error) {
	action, err := parseOperationString(line[3])
	if err != nil {
		return nil, errors.Wrap(err, "parsing operation string")
	}

	s := new(stock.Stock)
	if action != operation.Connectivity && action != operation.Interest {
		s, err = stockFinder.FindByName(line[2])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("find stock %s: %s", line[2], err.Error()))
		}
	}

	date := parseOperationDateString(line[1])
	amount, _ := strconv.Atoi(line[4])

	price := mm.Value{Amount: parseOperationPriceString(line[5])}
	priceChange := mm.Value{Amount: parseOperationPriceString(line[6])}
	priceChangeCommission := mm.Value{Amount: parseOperationPriceString(line[7]), Currency: mm.Euro}
	value := mm.Value{Amount: parseOperationPriceString(line[8]), Currency: mm.Euro}
	commission := mm.Value{Amount: parseOperationPriceString(line[9]), Currency: mm.Euro}

	o := operation.NewOperation(date, s, action, amount, price, priceChange, priceChangeCommission, value, commission)

	return o, nil
}

// parseOperationString - parse a potentially partial date string to Time
func parseOperationString(o string) (operation.Action, error) {
	if o == "" {
		return operation.Action(""), errors.New("operation can not be empty")
	}

	switch o {
	case "Compra":
		return operation.Buy, nil
	case "Venta":
		return operation.Sell, nil
	case "Conectividad":
		return operation.Connectivity, nil
	case "Dividendo":
		return operation.Dividend, nil
	case "Interés":
		return operation.Interest, nil
	}

	return operation.Action(""), errors.New("operation not valid")
}

// parseOperationDateString - parse a potentially partial date string to Time
func parseOperationDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parseOperationPriceString - parse a potentially partial date string to Time
func parseOperationPriceString(price string) float64 {
	price = strings.Replace(price, ",", ".", 1)

	p, _ := strconv.ParseFloat(price, 64)

	return p
}

func loadWalletWithActiveWalletItems(walletFinder wallet.Finder, stockFinder stock.Finder, name string) (*wallet.Wallet, error) {
	w, err := walletFinder.FindByName(name)
	if err != nil {
		return nil, err
	}

	if err = walletFinder.LoadActiveItems(w); err != nil {
		return nil, err
	}

	for _, i := range w.Items {
		// Add this into go routing. Use the example explain in the page
		// https://medium.com/@trevor4e/learning-gos-concurrency-through-illustrations-8c4aff603b3
		stk, err := stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			return nil, err
		}

		// I like to keep the address but change the content to keep,
		// trade and item pointing to the same stock
		*i.Stock = *stk
	}

	return w, err
}
