package dollarcostaverage

import (
	"errors"
	"testing"
	"time"

	"github.com/openware/gocryptotrader/backtester/common"
	"github.com/openware/gocryptotrader/backtester/data"
	"github.com/openware/gocryptotrader/backtester/data/kline"
	"github.com/openware/gocryptotrader/backtester/eventhandlers/strategies/base"
	"github.com/openware/gocryptotrader/backtester/eventtypes/event"
	eventkline "github.com/openware/gocryptotrader/backtester/eventtypes/kline"
	"github.com/openware/gocryptotrader/backtester/eventtypes/signal"
	"github.com/openware/pkg/asset"
	"github.com/openware/pkg/currency"
	gctkline "github.com/openware/pkg/kline"
	gctorder "github.com/openware/pkg/order"
)

func TestName(t *testing.T) {
	d := Strategy{}
	n := d.Name()
	if n != Name {
		t.Errorf("expected %v", Name)
	}
}

func TestSupportsSimultaneousProcessing(t *testing.T) {
	s := Strategy{}
	if !s.SupportsSimultaneousProcessing() {
		t.Error("expected true")
	}
}

func TestSetCustomSettings(t *testing.T) {
	s := Strategy{}
	err := s.SetCustomSettings(nil)
	if !errors.Is(err, base.ErrCustomSettingsUnsupported) {
		t.Errorf("expected: %v, received %v", base.ErrCustomSettingsUnsupported, err)
	}
}

func TestOnSignal(t *testing.T) {
	s := Strategy{}
	_, err := s.OnSignal(nil, nil)
	if !errors.Is(err, common.ErrNilEvent) {
		t.Errorf("expected: %v, received %v", common.ErrNilEvent, err)
	}

	dStart := time.Date(2020, 1, 0, 0, 0, 0, 0, time.UTC)
	dInsert := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dEnd := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	exch := "binance"
	a := asset.Spot
	p := currency.NewPair(currency.BTC, currency.USDT)
	d := data.Base{}
	d.SetStream([]common.DataEventHandler{&eventkline.Kline{
		Base: event.Base{
			Exchange:     exch,
			Time:         dInsert,
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
	da := &kline.DataFromKline{
		Item:  gctkline.Item{},
		Base:  d,
		Range: gctkline.IntervalRangeHolder{},
	}
	var resp signal.Event
	resp, err = s.OnSignal(da, nil)
	if err != nil {
		t.Error(err)
	}
	if resp.GetDirection() != common.MissingData {
		t.Error("expected missing data")
	}

	da.Item = gctkline.Item{
		Exchange: exch,
		Pair:     p,
		Asset:    a,
		Interval: gctkline.OneDay,
		Candles: []gctkline.Candle{
			{
				Time:   dInsert,
				Open:   1337,
				High:   1337,
				Low:    1337,
				Close:  1337,
				Volume: 1337,
			},
		},
	}
	err = da.Load()
	if err != nil {
		t.Error(err)
	}

	ranger := gctkline.CalculateCandleDateRanges(dStart, dEnd, gctkline.OneDay, 100000)
	da.Range = ranger
	_ = da.Range.VerifyResultsHaveData(da.Item.Candles)
	resp, err = s.OnSignal(da, nil)
	if err != nil {
		t.Error(err)
	}
	if resp.GetDirection() != gctorder.Buy {
		t.Errorf("expected buy, received %v", resp.GetDirection())
	}
}

func TestOnSignals(t *testing.T) {
	s := Strategy{}
	_, err := s.OnSignal(nil, nil)
	if !errors.Is(err, common.ErrNilEvent) {
		t.Errorf("expected: %v, received %v", common.ErrNilEvent, err)
	}
	dStart := time.Date(2020, 1, 0, 0, 0, 0, 0, time.UTC)
	dInsert := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dEnd := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	exch := "binance"
	a := asset.Spot
	p := currency.NewPair(currency.BTC, currency.USDT)
	d := data.Base{}
	d.SetStream([]common.DataEventHandler{&eventkline.Kline{
		Base: event.Base{
			Offset:       1,
			Exchange:     exch,
			Time:         dInsert,
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
	da := &kline.DataFromKline{
		Item:  gctkline.Item{},
		Base:  d,
		Range: gctkline.IntervalRangeHolder{},
	}
	var resp []signal.Event
	resp, err = s.OnSimultaneousSignals([]data.Handler{da}, nil)
	if err != nil {
		t.Error(err)
	}
	if len(resp) != 1 {
		t.Fatal("expected 1 response")
	}
	if resp[0].GetDirection() != common.MissingData {
		t.Error("expected missing data")
	}

	da.Item = gctkline.Item{
		Exchange: exch,
		Pair:     p,
		Asset:    a,
		Interval: gctkline.OneDay,
		Candles: []gctkline.Candle{
			{
				Time:   dInsert,
				Open:   1337,
				High:   1337,
				Low:    1337,
				Close:  1337,
				Volume: 1337,
			},
		},
	}
	err = da.Load()
	if err != nil {
		t.Error(err)
	}

	ranger := gctkline.CalculateCandleDateRanges(dStart, dEnd, gctkline.OneDay, 100000)
	da.Range = ranger
	_ = da.Range.VerifyResultsHaveData(da.Item.Candles)
	resp, err = s.OnSimultaneousSignals([]data.Handler{da}, nil)
	if err != nil {
		t.Error(err)
	}
	if len(resp) != 1 {
		t.Fatal("expected 1 response")
	}
	if resp[0].GetDirection() != gctorder.Buy {
		t.Error("expected buy")
	}
}

func TestSetDefaults(t *testing.T) {
	s := Strategy{}
	s.SetDefaults()
	if s != (Strategy{}) {
		t.Error("expected no changes")
	}
}
