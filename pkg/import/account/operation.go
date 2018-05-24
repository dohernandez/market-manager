package import_account

import (
	"context"
	"time"

	"io"

	"strconv"

	"strings"

	"fmt"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	ImportAccount struct {
		ctx    context.Context
		reader _import.Reader

		purchaseService *purchase.Service
		accountService  *account.Service
	}
)

func NewImportAccount(
	ctx context.Context,
	reader _import.Reader,
	purchaseService *purchase.Service,
	accountService *account.Service,
) *ImportAccount {
	return &ImportAccount{
		ctx:             ctx,
		reader:          reader,
		purchaseService: purchaseService,
		accountService:  accountService,
	}
}

func (i *ImportAccount) Import() error {
	i.reader.Open()
	defer i.reader.Close()

	name := i.ctx.Value("wallet").(string)
	if name == "" {
		return errors.New("missing wallet name")
	}

	w, err := i.accountService.FindWalletByName(name)
	if err != nil {
		return err
	}

	for {
		line, err := i.reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(i.ctx).Fatal(err)
		}

		action, err := i.parseOperationString(line[3])
		if err != nil {
			return err
		}

		s := new(stock.Stock)
		if action != operation.Connectivity && action != operation.Interest {
			s, err = i.purchaseService.FindStockByName(line[2])
			if err != nil {
				return errors.New(fmt.Sprintf("find stock %s: %s", line[2], err.Error()))
			}
		}

		date := i.parseDateString(line[1])
		amount, _ := strconv.Atoi(line[4])

		price := mm.Value{Amount: i.parsePriceString(line[5])}
		priceChange := mm.Value{Amount: i.parsePriceString(line[6])}
		priceChangeCommission := mm.Value{Amount: i.parsePriceString(line[7]), Currency: mm.Euro}
		value := mm.Value{Amount: i.parsePriceString(line[8]), Currency: mm.Euro}
		commission := mm.Value{Amount: i.parsePriceString(line[9]), Currency: mm.Euro}

		o := operation.NewOperation(date, s, action, amount, price, priceChange, priceChangeCommission, value, commission)

		w.AddOperation(o)
	}

	return i.accountService.SaveAllOperations(w)
}

// parseDateString - parse a potentially partial date string to Time
func (i *ImportAccount) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parseOperationString - parse a potentially partial date string to Time
func (i *ImportAccount) parseOperationString(o string) (operation.Action, error) {
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
func (i *ImportAccount) parsePriceString(price string) float64 {
	price = strings.Replace(price, ",", ".", 1)

	p, _ := strconv.ParseFloat(price, 64)

	return p
}
