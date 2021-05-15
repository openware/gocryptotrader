package exchange

import (
	"errors"

	"github.com/openware/gocryptotrader/backtester/config"
	"github.com/openware/gocryptotrader/backtester/data"
	"github.com/openware/gocryptotrader/backtester/eventtypes/fill"
	"github.com/openware/gocryptotrader/backtester/eventtypes/order"
	"github.com/openware/gocryptotrader/engine"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/currency"
	gctorder "github.com/openware/pkg/order"
)

var (
	errDataMayBeIncorrect = errors.New("data may be incorrect")
)

// ExecutionHandler interface dictates what functions are required to submit an order
type ExecutionHandler interface {
	SetExchangeAssetCurrencySettings(string, asset.Item, currency.Pair, *Settings)
	GetCurrencySettings(string, asset.Item, currency.Pair) (Settings, error)
	ExecuteOrder(order.Event, data.Handler, *engine.Engine) (*fill.Fill, error)
	Reset()
}

// Exchange contains all the currency settings
type Exchange struct {
	CurrencySettings []Settings
}

// Settings allow the eventhandler to size an order within the limitations set by the config file
type Settings struct {
	ExchangeName  string
	UseRealOrders bool

	InitialFunds float64

	CurrencyPair currency.Pair
	AssetType    asset.Item

	ExchangeFee float64
	MakerFee    float64
	TakerFee    float64

	BuySide  config.MinMax
	SellSide config.MinMax

	Leverage config.Leverage

	MinimumSlippageRate float64
	MaximumSlippageRate float64

	Limits               *gctorder.Limits
	CanUseExchangeLimits bool
}
