package import_account

import (
	"context"
	"io"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application/import"
	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type (
	ImportWallet struct {
		ctx    context.Context
		reader _import.Reader

		accountService *service.Account
		bankingService *service.Banking
	}
)

var _ _import.Import = &ImportWallet{}

func NewImportWallet(
	ctx context.Context,
	reader _import.Reader,
	accountService *service.Account,
	bankingService *service.Banking,
) *ImportWallet {
	return &ImportWallet{
		ctx:            ctx,
		reader:         reader,
		accountService: accountService,
		bankingService: bankingService,
	}
}

func (i *ImportWallet) Import() error {
	i.reader.Open()
	defer i.reader.Close()

	name := i.ctx.Value("wallet").(string)
	if name == "" {
		return errors.New("missing wallet name")
	}

	var ws []*wallet.Wallet

	for {
		line, err := i.reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(i.ctx).Fatal(err)
		}

		url := line[0]
		bankAccount, err := i.bankingService.FindBankAccountByAlias(line[1])
		if err != nil {
			return err
		}

		w := wallet.NewWallet(name, url)
		err = w.AddBankAccount(bankAccount)
		if err != nil {
			return err
		}

		ws = append(ws, w)
	}

	return i.accountService.SaveAllWallets(ws)
}