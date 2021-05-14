package settings

import (
	"github.com/openware/gocryptotrader/backtester/config"
	"github.com/openware/gocryptotrader/backtester/eventhandlers/portfolio/compliance"
	"github.com/openware/gocryptotrader/backtester/eventhandlers/portfolio/holdings"
)

// Settings holds all important information for the portfolio manager
// to assess purchasing decisions
type Settings struct {
	InitialFunds      float64
	Fee               float64
	BuySideSizing     config.MinMax
	SellSideSizing    config.MinMax
	Leverage          config.Leverage
	HoldingsSnapshots []holdings.Holding
	ComplianceManager compliance.Manager
}
