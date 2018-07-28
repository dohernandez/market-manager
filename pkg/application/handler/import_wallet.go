package handler

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
)

type importWallet struct {
	bankAccountFinder bank.Finder
	walletPersister   wallet.Persister
}

func NewImportWallet(
	bankAccountFinder bank.Finder,
	walletPersister wallet.Persister,
) *importWallet {
	return &importWallet{
		bankAccountFinder: bankAccountFinder,
		walletPersister:   walletPersister,
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

	err = h.walletPersister.PersistAll(ws)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting wallets -> error [%s]",
			err,
		)

		return nil, err
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
