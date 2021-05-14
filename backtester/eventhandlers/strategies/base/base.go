package base

import (
	"github.com/openware/gocryptotrader/backtester/common"
	"github.com/openware/gocryptotrader/backtester/data"
	"github.com/openware/gocryptotrader/backtester/eventtypes/event"
	"github.com/openware/gocryptotrader/backtester/eventtypes/signal"
)

// Strategy is base implementation of the Handler interface
type Strategy struct {
	useSimultaneousProcessing bool
}

// GetBaseData returns the non-interface version of the Handler
func (s *Strategy) GetBaseData(d data.Handler) (signal.Signal, error) {
	if d == nil {
		return signal.Signal{}, common.ErrNilArguments
	}
	latest := d.Latest()
	if latest == nil {
		return signal.Signal{}, common.ErrNilEvent
	}
	return signal.Signal{
		Base: event.Base{
			Offset:       latest.GetOffset(),
			Exchange:     latest.GetExchange(),
			Time:         latest.GetTime(),
			CurrencyPair: latest.Pair(),
			AssetType:    latest.GetAssetType(),
			Interval:     latest.GetInterval(),
		},
		ClosePrice: latest.ClosePrice(),
		HighPrice:  latest.HighPrice(),
		OpenPrice:  latest.OpenPrice(),
		LowPrice:   latest.LowPrice(),
	}, nil
}

// UseSimultaneousProcessing returns whether multiple currencies can be assessed in one go
func (s *Strategy) UseSimultaneousProcessing() bool {
	return s.useSimultaneousProcessing
}

// SetSimultaneousProcessing sets whether multiple currencies can be assessed in one go
func (s *Strategy) SetSimultaneousProcessing(b bool) {
	s.useSimultaneousProcessing = b
}
