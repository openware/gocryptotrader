package exchangerates

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/openware/gocryptotrader/common"
	"github.com/openware/gocryptotrader/log"
	"github.com/openware/pkg/currency/forexprovider/base"
	"github.com/openware/pkg/request"
)

// Setup sets appropriate values for CurrencyLayer
func (e *ExchangeRates) Setup(config base.Settings) error {
	if config.APIKey == "" {
		return errors.New("API key must be set")
	}
	e.Name = config.Name
	e.Enabled = config.Enabled
	e.RESTPollingDelay = config.RESTPollingDelay
	e.Verbose = config.Verbose
	e.PrimaryProvider = config.PrimaryProvider
	e.APIKey = config.APIKey
	e.APIKeyLvl = config.APIKeyLvl
	e.Requester = request.New(e.Name,
		common.NewHTTPClientWithTimeout(base.DefaultTimeOut),
		request.WithLimiter(request.NewBasicRateLimit(rateLimitInterval, requestRate)))
	return nil
}

func (e *ExchangeRates) cleanCurrencies(baseCurrency, symbols string) string {
	if len(e.supportedCurrencies) == 0 {
		supportedCurrencies, err := e.GetSupportedCurrencies()
		if err != nil {
			log.Warnf(log.Global, "ExchangeRatesAPI unable to fetch supported currencies: %s", err)
		} else {
			e.supportedCurrencies = supportedCurrencies
		}
	}
	var cleanedCurrencies []string
	symbols = strings.Replace(symbols, "RUR", "RUB", -1)
	var s = strings.Split(symbols, ",")
	for _, x := range s {
		// first make sure that the baseCurrency is not in the symbols list
		// if it is set
		if baseCurrency != "" {
			if x == baseCurrency {
				continue
			}
		} else {
			// otherwise since the baseCurrency is empty, make sure that it
			// does not exist in the symbols list
			if x == "EUR" {
				continue
			}
		}

		// remove and warn about any unsupported currencies
		if len(e.supportedCurrencies) > 0 {
			if !strings.Contains(strings.Join(e.supportedCurrencies, ","), x) {
				log.Warnf(log.Global,
					"Forex provider ExchangeRatesAPI does not support currency %s, removing from forex rates query.\n", x)
				continue
			}
		}
		cleanedCurrencies = append(cleanedCurrencies, x)
	}
	return strings.Join(cleanedCurrencies, ",")
}

// GetSymbols returns a list of supported symbols
func (e *ExchangeRates) GetSymbols() (map[string]string, error) {
	resp := struct {
		Symbols map[string]string `json:"symbols"`
	}{}
	return resp.Symbols, e.SendHTTPRequest("symbols", url.Values{}, &resp)
}

// GetLatestRates returns a map of forex rates based on the supplied params
// baseCurrency - USD	[optional] The base currency to use for forex rates, defaults to EUR
// symbols - AUD,USD	[optional] The symbols to query the forex rates for, default is
// all supported currencies
func (e *ExchangeRates) GetLatestRates(baseCurrency, symbols string) (*Rates, error) {
	vals := url.Values{}
	if len(baseCurrency) > 0 && e.APIKeyLvl <= apiKeyFree && !strings.EqualFold("EUR", baseCurrency) {
		return nil, errCannotSetBaseCurrencyOnFreePlan
	} else if len(baseCurrency) > 0 {
		vals.Set("base", baseCurrency)
	}

	if len(symbols) > 0 {
		symbols = e.cleanCurrencies(baseCurrency, symbols)
		vals.Set("symbols", symbols)
	}

	var result Rates
	return &result, e.SendHTTPRequest(exchangeRatesLatest, vals, &result)
}

// GetHistoricalRates returns historical exchange rate data for all available or
// a specific set of currencies.
// date - YYYY-MM-DD	[required] A date in the past
// baseCurrency - USD 			[optional] The base currency to use for forex rates, defaults to EUR
// symbols - AUD,USD	[optional] The symbols to query the forex rates for, default is
// all supported currencies
func (e *ExchangeRates) GetHistoricalRates(date time.Time, baseCurrency string, symbols []string) (*HistoricalRates, error) {
	if date.IsZero() {
		return nil, errors.New("a date must be specified")
	}

	var resp HistoricalRates
	v := url.Values{}

	if len(baseCurrency) > 0 && e.APIKeyLvl <= apiKeyFree && !strings.EqualFold("EUR", baseCurrency) {
		return nil, errCannotSetBaseCurrencyOnFreePlan
	} else if len(baseCurrency) > 0 {
		v.Set("base", baseCurrency)
	}

	if len(symbols) > 0 {
		s := e.cleanCurrencies(baseCurrency, strings.Join(symbols, ","))
		v.Set("symbols", s)
	}

	return &resp, e.SendHTTPRequest(date.UTC().Format(timeLayout), v, &resp)
}

// ConvertCurrency converts a currency based on the supplied params
func (e *ExchangeRates) ConvertCurrency(from, to string, amount float64, date time.Time) (*ConvertCurrency, error) {
	if e.APIKeyLvl <= apiKeyFree {
		return nil, errAPIKeyLevelRestrictedAccess
	}
	vals := url.Values{}
	if from == "" || to == "" || amount == 0 {
		return nil, errors.New("from, to and amount must be set")
	}

	vals.Set("from", from)
	vals.Set("to", to)
	vals.Set("amount", strconv.FormatFloat(amount, 'e', -1, 64))

	if !date.IsZero() {
		vals.Set("date", date.UTC().Format(timeLayout))
	}

	var cc ConvertCurrency
	return &cc, e.SendHTTPRequest(exchangeRatesConvert, vals, &cc)
}

// GetTimeSeriesRates returns daily historical exchange rate data between two
// specified dates for all available or a specific set of currencies.
// startDate - YYYY-MM-DD	[required] A date in the past
// endDate - YYYY-MM-DD	[required] A date in the past but greater than the startDate
// baseCurrency - USD 	[optional] The base currency to use for forex rates, defaults to EUR
// symbols - AUD,USD 	[optional] The symbols to query the forex rates for, default is
// all supported currencies
func (e *ExchangeRates) GetTimeSeriesRates(startDate, endDate time.Time, baseCurrency string, symbols []string) (*TimeSeriesRates, error) {
	if e.APIKeyLvl <= apiKeyFree {
		return nil, errAPIKeyLevelRestrictedAccess
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, errors.New("startDate and endDate params must be set")
	}

	if startDate.After(endDate) {
		return nil, errors.New("startDate must be before endDate")
	}

	v := url.Values{}
	v.Set("start_date", startDate.UTC().Format(timeLayout))
	v.Set("end_date", endDate.UTC().Format(timeLayout))

	if baseCurrency != "" {
		v.Set("base", baseCurrency)
	}

	if len(symbols) > 0 {
		s := e.cleanCurrencies(baseCurrency, strings.Join(symbols, ","))
		v.Set("symbols", s)
	}

	var resp TimeSeriesRates
	return &resp, e.SendHTTPRequest(exchangeRatesTimeSeries, v, &resp)
}

// GetFluctuations returns rate fluctuations based on the supplied params
func (e *ExchangeRates) GetFluctuations(startDate, endDate time.Time, baseCurrency, symbols string) (*Fluctuations, error) {
	if e.APIKeyLvl <= apiKeyFree {
		return nil, errAPIKeyLevelRestrictedAccess
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, errors.New("startDate and endDate must be set")
	}

	if startDate.After(endDate) {
		return nil, errors.New("startDate must be before endDate")
	}

	v := url.Values{}
	v.Set("start_date", startDate.UTC().Format(timeLayout))
	v.Set("end_date", endDate.UTC().Format(timeLayout))

	if baseCurrency != "" {
		v.Set("base", baseCurrency)
	}
	if symbols != "" {
		v.Set("symbols", symbols)
	}

	var f Fluctuations
	return &f, e.SendHTTPRequest(exchangeRatesFluctuation, v, &f)
}

// GetRates is a wrapper function to return forex rates
func (e *ExchangeRates) GetRates(baseCurrency, symbols string) (map[string]float64, error) {
	result, err := e.GetLatestRates(baseCurrency, symbols)
	if err != nil {
		return nil, err
	}

	standardisedRates := make(map[string]float64)
	for k, v := range result.Rates {
		curr := baseCurrency + k
		standardisedRates[curr] = v
	}

	return standardisedRates, nil
}

// GetSupportedCurrencies returns the supported currency list
func (e *ExchangeRates) GetSupportedCurrencies() ([]string, error) {
	symbols, err := e.GetSymbols()
	if err != nil {
		return nil, err
	}

	var supportedCurrencies []string
	for x := range symbols {
		supportedCurrencies = append(supportedCurrencies, x)
	}
	e.supportedCurrencies = supportedCurrencies
	return supportedCurrencies, nil
}

// SendHTTPRequest sends a HTTPS request to the desired endpoint and returns the result
func (e *ExchangeRates) SendHTTPRequest(endPoint string, values url.Values, result interface{}) error {
	if e.APIKey == "" {
		return errors.New("api key must be set")
	}
	values.Set("access_key", e.APIKey)
	protocolScheme := "https://"
	if e.APIKeyLvl == apiKeyFree {
		protocolScheme = "http://"
	}
	path := common.EncodeURLValues(protocolScheme+exchangeRatesAPI+"/v1/"+endPoint, values)
	err := e.Requester.SendPayload(context.Background(), &request.Item{
		Method:  http.MethodGet,
		Path:    path,
		Result:  result,
		Verbose: e.Verbose,
	})
	if err != nil {
		return fmt.Errorf("exchangeRatesAPI: SendHTTPRequest error %s with path %s",
			err,
			path)
	}
	return nil
}
