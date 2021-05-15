package base

import (
	"errors"
	"testing"
	"time"

	"github.com/openware/gocryptotrader/backtester/common"
	"github.com/openware/gocryptotrader/backtester/data"
	datakline "github.com/openware/gocryptotrader/backtester/data/kline"
	"github.com/openware/gocryptotrader/backtester/eventtypes/event"
	"github.com/openware/gocryptotrader/backtester/eventtypes/kline"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/currency"
	gctkline "github.com/openware/pkg/kline"
)

func TestGetBase(t *testing.T) {
	s := Strategy{}
	_, err := s.GetBaseData(nil)
	if !errors.Is(err, common.ErrNilArguments) {
		t.Errorf("expected: %v, received %v", common.ErrNilArguments, err)
	}

	_, err = s.GetBaseData(&datakline.DataFromKline{})
	if !errors.Is(err, common.ErrNilEvent) {
		t.Errorf("expected: %v, received %v", common.ErrNilEvent, err)
	}
	tt := time.Now()
	exch := "binance"
	a := asset.Spot
	p := currency.NewPair(currency.BTC, currency.USDT)
	d := data.Base{}
	d.SetStream([]common.DataEventHandler{&kline.Kline{
		Base: event.Base{
			Exchange:     exch,
			Time:         tt,
			Interval:     gctkline.OneDay,
			CurrencyPair: p,
			AssetType:    a,
		},
		Open:   1337,
		Close:  1337,
		Low:    1337,
		High:   1337,
		Volume: 1337,
	}})

	d.Next()
	_, err = s.GetBaseData(&datakline.DataFromKline{
		Item:  gctkline.Item{},
		Base:  d,
		Range: gctkline.IntervalRangeHolder{},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSetSimultaneousProcessing(t *testing.T) {
	s := Strategy{}
	is := s.UseSimultaneousProcessing()
	if is {
		t.Error("expected false")
	}
	s.SetSimultaneousProcessing(true)
	is = s.UseSimultaneousProcessing()
	if !is {
		t.Error("expected true")
	}
}
