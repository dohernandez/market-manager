package handler

import (
	"context"
	"errors"
	"io"

	"github.com/gogolfing/cbus"

	"time"

	"strconv"
	"strings"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
)

type importWallet struct {
	bankAccountFinder bank.Finder
}

func NewImportWallet(
	bankAccountFinder bank.Finder,
) *importWallet {
	return &importWallet{
		bankAccountFinder: bankAccountFinder,
	}
}

func (h *importWallet) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	filePath := command.(*appCommand.ImportWallet).FilePath
	r := util.NewCsvReader(filePath)

	r.Open()
	defer r.Close()

	name := command.(*appCommand.ImportWallet).Name
	if name == "" {
		return nil, errors.New("missing wallet name")
	}

	var ws []*wallet.Wallet
	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(ctx).Fatal(err)
		}

		url := line[0]
		bankAccount, err := h.bankAccountFinder.FindByAlias(line[1])
		if err != nil {
			return nil, err
		}

		w := wallet.NewWallet(name, url)
		err = w.AddBankAccount(bankAccount)
		if err != nil {
			return nil, err
		}

		ws = append(ws, w)
	}

	return ws, nil
}

// parseDateString - parse a potentially partial date string to Time
func (h *importWallet) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parsePriceString - parse a potentially float string to float64
func (h *importWallet) parsePriceString(price string) (float64, error) {
	price = strings.Replace(price, ".", "", 1)
	price = strings.Replace(price, ",", ".", 1)

	return strconv.ParseFloat(price, 64)
}
