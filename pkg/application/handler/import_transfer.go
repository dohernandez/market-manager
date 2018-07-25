package handler

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gogolfing/cbus"

	"time"

	"strconv"
	"strings"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type importTransfer struct {
	bankAccountFinder bank.Finder
}

func NewImportTransfer(
	bankAccountFinder bank.Finder,
) *importTransfer {
	return &importTransfer{
		bankAccountFinder: bankAccountFinder,
	}
}

func (h *importTransfer) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	filePath := command.(*appCommand.ImportTransfer).FilePath
	r := util.NewCsvReader(filePath)

	r.Open()
	defer r.Close()

	var ts []*transfer.Transfer
	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(ctx).Fatal(err)
		}

		date := h.parseDateString(line[0])

		from, err := h.bankAccountFinder.FindByAlias(line[1])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s %q", err.Error(), line[1]))
		}

		to, err := h.bankAccountFinder.FindByAlias(line[2])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s %q", err.Error(), line[2]))
		}

		amount, err := h.parsePriceString(line[3])
		if err != nil {
			return nil, err
		}

		t := transfer.NewTransfer(from, to, amount, date)

		ts = append(ts, t)
	}

	return ts, nil
}

// parseDateString - parse a potentially partial date string to Time
func (h *importTransfer) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parsePriceString - parse a potentially float string to float64
func (h *importTransfer) parsePriceString(price string) (float64, error) {
	price = strings.Replace(price, ".", "", 1)
	price = strings.Replace(price, ",", ".", 1)

	return strconv.ParseFloat(price, 64)
}
