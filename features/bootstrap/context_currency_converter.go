package bootstrap

import (
	"net/http"

	"fmt"

	"encoding/json"

	"strconv"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

type currencyConverterContext struct {
	currencyConverterAPI *WireMock
}

func RegisterCurrencyConverterContext(s *godog.Suite, currencyConverterAPI *WireMock) {
	ccc := &currencyConverterContext{
		currencyConverterAPI: currencyConverterAPI,
	}

	s.Step(`^when request currency converter:$`, ccc.whenRequestCurrencyConverter)
}

func (c *currencyConverterContext) whenRequestCurrencyConverter(currencyConverters *gherkin.DataTable) error {

	for _, currencyConverter := range currencyConverters.Rows[1:] {
		wmRequest := NewWireMockRequest()
		wmRequest.SetMethod(http.MethodGet)
		wmRequest.SetURL(fmt.Sprintf("/api/v5/convert?q=%s&compact=ultra", currencyConverter.Cells[0].Value))
		c.currencyConverterAPI.SetWireMockRequest(wmRequest)

		vcc, err := strconv.ParseFloat(currencyConverter.Cells[1].Value, 64)
		if err != nil {
			return err
		}

		wmResponse := NewWireMockResponse()

		if currencyConverter.Cells[0].Value == "EUR_USD" {

			body, err := json.Marshal(struct {
				EURUSD float64 `json:"EUR_USD"`
			}{
				EURUSD: vcc,
			})
			if err != nil {
				return err
			}

			wmResponse.SetBody(string(body))
		} else {
			body, err := json.Marshal(struct {
				EURCAD float64 `json:"EUR_CAD"`
			}{
				EURCAD: vcc,
			})
			if err != nil {
				return err
			}

			wmResponse.SetBody(string(body))
		}

		wmResponse.SetHeader("Content-Type", "application/json")
		wmResponse.SetStatus(http.StatusOK)
		c.currencyConverterAPI.SetWireMockResponse(wmResponse)

		err = c.currencyConverterAPI.Send()
		if err != nil {
			return err
		}
	}

	return nil
}
