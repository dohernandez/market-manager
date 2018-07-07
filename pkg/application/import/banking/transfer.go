package import_banking

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type (
	ImportTransfer struct {
		ctx    context.Context
		reader _import.Reader

		bankingService *service.Banking
	}
)

var _ _import.Import = &ImportTransfer{}

func NewImportTransfer(
	ctx context.Context,
	reader _import.Reader,
	bankingService *service.Banking,
) *ImportTransfer {
	return &ImportTransfer{
		ctx:            ctx,
		reader:         reader,
		bankingService: bankingService,
	}
}

func (i *ImportTransfer) Import() error {
	i.reader.Open()
	defer i.reader.Close()

	var ts []*transfer.Transfer
	for {
		line, err := i.reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(i.ctx).Fatal(err)
		}

		date := i.parseDateString(line[0])

		from, err := i.bankingService.FindBankAccountByAlias(line[1])
		if err != nil {
			return errors.New(fmt.Sprintf("%s %q", err.Error(), line[1]))
		}

		to, err := i.bankingService.FindBankAccountByAlias(line[2])
		if err != nil {
			return errors.New(fmt.Sprintf("%s %q", err.Error(), line[2]))
		}

		amount, err := i.parsePriceString(line[3])
		if err != nil {
			return err
		}

		t := transfer.NewTransfer(from, to, amount, date)

		ts = append(ts, t)
	}

	return i.bankingService.SaveAllTransfers(ts)
}

// parseDateString - parse a potentially partial date string to Time
func (i *ImportTransfer) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parsePriceString - parse a potentially float string to float64
func (i *ImportTransfer) parsePriceString(price string) (float64, error) {
	price = strings.Replace(price, ".", "", 1)
	price = strings.Replace(price, ",", ".", 1)

	return strconv.ParseFloat(price, 64)
}
