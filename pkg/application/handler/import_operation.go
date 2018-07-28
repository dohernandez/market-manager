package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type importOperation struct {
	stockFinder     stock.Finder
	walletFinder    wallet.Finder
	walletPersister wallet.Persister
	ccClient        *cc.Client
}

func NewImportOperation(
	stockFinder stock.Finder,
	walletFinder wallet.Finder,
	walletPersister wallet.Persister,
	ccClient *cc.Client,
) *importOperation {
	return &importOperation{
		stockFinder:     stockFinder,
		walletFinder:    walletFinder,
		walletPersister: walletPersister,
		ccClient:        ccClient,
	}
}

func (h *importOperation) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	filePath := command.(*appCommand.ImportOperation).FilePath
	r := util.NewCsvReader(filePath)

	r.Open()
	defer r.Close()

	wName := command.(*appCommand.ImportOperation).Wallet
	if wName == "" {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet -> error [%s]",
			err,
		)

		return nil, errors.New("missing wallet name")
	}

	w, err := h.walletFinder.FindByName(wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet name [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	err = h.LoadActiveWalletItems(w)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet active items [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	var os []*operation.Operation

	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(ctx).Fatal(err)
		}

		action, err := h.parseOperationString(line[3])
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while parseOperationString [%s] -> error [%s]",
				line[3],
				err,
			)

			return nil, err
		}

		s := new(stock.Stock)
		if action != operation.Connectivity && action != operation.Interest {
			s, err = h.stockFinder.FindByName(line[2])
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while finding stock by name [%s] -> error [%s]",
					line[2],
					err,
				)

				return nil, errors.New(fmt.Sprintf("find stock %s: %s", line[2], err.Error()))
			}
		}

		date := h.parseDateString(line[1])
		amount, _ := strconv.Atoi(line[4])

		price := mm.Value{Amount: h.parsePriceString(line[5])}
		priceChange := mm.Value{Amount: h.parsePriceString(line[6])}
		priceChangeCommission := mm.Value{Amount: h.parsePriceString(line[7]), Currency: mm.Euro}
		value := mm.Value{Amount: h.parsePriceString(line[8]), Currency: mm.Euro}
		commission := mm.Value{Amount: h.parsePriceString(line[9]), Currency: mm.Euro}

		o := operation.NewOperation(date, s, action, amount, price, priceChange, priceChangeCommission, value, commission)

		os = append(os, o)
		w.AddOperation(o)
	}

	err = h.walletPersister.PersistOperations(w)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting operation -> error [%s]",
			err,
		)

		return nil, err
	}

	return os, nil
}

func (h *importOperation) LoadActiveWalletItems(w *wallet.Wallet) error {
	err := h.walletFinder.LoadActiveItems(w)
	if err != nil {
		return err
	}

	cEURUSD, err := h.ccClient.Converter.Get()
	if err != nil {
		return err
	}

	for _, i := range w.Items {
		// Add this into go routing. Use the example explain in the page
		// https://medium.com/@trevor4e/learning-gos-concurrency-through-illustrations-8c4aff603b3
		stk, err := h.stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			return err
		}

		i.CapitalRate = cEURUSD.EURUSD
		i.Stock = stk
	}

	return nil
}

// parseDateString - parse a potentially partial date string to Time
func (h *importOperation) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parseOperationString - parse a potentially partial date string to Time
func (h *importOperation) parseOperationString(o string) (operation.Action, error) {
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

// parseDateString - parse a potentially partial date string to Time
func (h *importOperation) parsePriceString(price string) float64 {
	price = strings.Replace(price, ",", ".", 1)

	p, _ := strconv.ParseFloat(price, 64)

	return p
}
