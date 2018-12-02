package config

import (
	"github.com/kelseyhightower/envconfig"

	"github.com/dohernandez/go-quote"
)

// Specification structured configuration variables.
type Specification struct {
	Debug       bool   `envconfig:"DEBUG" default:"false"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	Environment string `envconfig:"ENVIRONMENT" default:"staging"`

	Database struct {
		DSN            string `envconfig:"PGSQL_DSN"`
		MigrationsPath string `envconfig:"MIGRATIONS_PATH" default:"file://resources/migrations"`
	}
	HTTP struct {
		BaseURL string `envconfig:"HTTP_BASE_URL"`
		Port    int    `envconfig:"HTTP_PORT" default:"8000"`
	}
	Import struct {
		AccountsPath   string `envconfig:"ACCOUNTS_PATH" default:"resources/import/accounts"`
		StocksPath     string `envconfig:"STOCKS_PATH" default:"resources/import/stocks"`
		TransfersPath  string `envconfig:"TRANSFERS_PATH" default:"resources/import/transfers"`
		WalletsPath    string `envconfig:"WALLETS_PATH" default:"resources/import/wallets"`
		RetentionsPath string `envconfig:"RETENTIONS_PATH" default:"resources/import/retentions"`
	}

	IEXTrading struct {
		Timeout int `envconfig:"IEX_TRADING_TIMEOUT" default:"30"`
	}
	CurrencyConverter struct {
		BaseURL string `envconfig:"CURRENCY_CONVERTER_BASEURL" default:"http://free.currencyconverterapi.com/"`
		Timeout int    `envconfig:"CURRENCY_CONVERTER_TIMEOUT" default:"15"`
	}
	QuoteScraper struct {
		FinanceYahooBaseURL  string `envconfig:"FINANCE_YAHOO_BASEURL" default:"https://finance.yahoo.com"`
		Query1YahooBaseURL   string `envconfig:"QUERY1_YAHOO_BASEURL" default:"https://query1.finance.yahoo.com"`
		FinanceYahooQuoteURL string `envconfig:"FINANCE_YAHOO_QUOTE_URL" default:"https://finance.yahoo.com/quote"`
		MarketChameleonURL   string `envconfig:"FINANCE_YAHOO_QUOTE_URL" default:"https://marketchameleon.com/Overview"`
		MarketChameleonPath  string `envconfig:"FINANCE_YAHOO_QUOTE_URL" default:"resources/import/market-chameleon"`
	}

	Degiro struct {
		Retention float64 `envconfig:"RETENTION" default:"15"`
		Exchanges struct {
			NASDAQ struct {
				Commission struct {
					Base struct {
						Amount   float64 `envconfig:"NASDAQ_COMMISSION_BASE" default:"0.50"`
						Currency string  `envconfig:"NASDAQ_COMMISSION_BASE_CURRENCY" default:"€"`
					}
					Extra struct {
						Amount   float64 `envconfig:"NASDAQ_COMMISSION_EXTRA" default:"0.004"`
						Currency string  `envconfig:"NASDAQ_COMMISSION_EXTRA_CURRENCY" default:"$"`
						Apply    string  `envconfig:"NASDAQ_COMMISSION_EXTRA_APPLY" default:"PER_STOCK"`
					}
				}
				ChangeCommission struct {
					Amount   float64 `envconfig:"NASDAQ_CHANGE_COMMISSION" default:"0.16"`
					Currency string  `envconfig:"NASDAQ_CHANGE_COMMISSION_CURRENCY" default:"€"`
				}
			}
			NYSE struct {
				Commission struct {
					Base struct {
						Amount   float64 `envconfig:"NYSE_COMMISSION_BASE" default:"0.50"`
						Currency string  `envconfig:"NYSE_COMMISSION_BASE_CURRENCY" default:"€"`
					}
					Extra struct {
						Amount   float64 `envconfig:"NYSE_COMMISSION_EXTRA" default:"0.004"`
						Currency string  `envconfig:"NYSE_COMMISSION_EXTRA_CURRENCY" default:"$"`
						Apply    string  `envconfig:"NYSE_COMMISSION_EXTRA_APPLY" default:"PER_STOCK"`
					}
				}
				ChangeCommission struct {
					Amount   float64 `envconfig:"NASDAQ_CHANGE_COMMISSION" default:"0.16"`
					Currency string  `envconfig:"NASDAQ_CHANGE_COMMISSION_CURRENCY" default:"€"`
				}
			}
			BME struct {
				Commission struct {
					Base struct {
						Amount   float64 `envconfig:"BME_COMMISSION_BASE" default:"2"`
						Currency string  `envconfig:"BME_COMMISSION_BASE_CURRENCY" default:"€"`
					}
					Extra struct {
						Amount   float64 `envconfig:"BME_COMMISSION_EXTRA" default:"0.04"`
						Currency string  `envconfig:"BME_COMMISSION_EXTRA_CURRENCY" default:"€"`
						Apply    string  `envconfig:"BME_COMMISSION_EXTRA_APPLY" default:"INVESTED_PERCENTAGE"`
					}
					Maximum struct {
						Amount   float64 `envconfig:"FRA_COMMISSION_MAXIMUM" default:"60"`
						Currency string  `envconfig:"FRA_COMMISSION_MAXIMUM_CURRENCY" default:"€"`
					}
				}
			}
			FRA struct {
				Commission struct {
					Base struct {
						Amount   float64 `envconfig:"FRA_COMMISSION_BASE" default:"7.5"`
						Currency string  `envconfig:"FRA_COMMISSION_BASE_CURRENCY" default:"€"`
					}
					Extra struct {
						Amount   float64 `envconfig:"FRA_COMMISSION_EXTRA" default:"0.08"`
						Currency string  `envconfig:"FRA_COMMISSION_EXTRA_CURRENCY" default:"€"`
						Apply    string  `envconfig:"FRA_COMMISSION_EXTRA_APPLY" default:"INVESTED_PERCENTAGE"`
					}
				}
			}
			BIT struct {
				Commission struct {
					Base struct {
						Amount   float64 `envconfig:"FRA_COMMISSION_BASE" default:"4"`
						Currency string  `envconfig:"FRA_COMMISSION_BASE_CURRENCY" default:"€"`
					}
					Extra struct {
						Amount   float64 `envconfig:"FRA_COMMISSION_EXTRA" default:"0.04"`
						Currency string  `envconfig:"FRA_COMMISSION_EXTRA_CURRENCY" default:"€"`
						Apply    string  `envconfig:"FRA_COMMISSION_EXTRA_APPLY" default:"INVESTED_PERCENTAGE"`
					}
					Maximum struct {
						Amount   float64 `envconfig:"FRA_COMMISSION_MAXIMUM" default:"60"`
						Currency string  `envconfig:"FRA_COMMISSION_MAXIMUM_CURRENCY" default:"€"`
					}
				}
			}
		}
	}
}

// LoadEnv load config variables into Specification.
func LoadEnv() (*Specification, error) {
	var config Specification
	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	quote.YahooUrls.FinanceBaseURL = config.QuoteScraper.FinanceYahooBaseURL
	quote.YahooUrls.Query1BaseURL = config.QuoteScraper.Query1YahooBaseURL

	return &config, err
}
