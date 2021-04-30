package opendax

import (
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type OrderMessage struct {
	Market         string
	ID             uint64
	UUID           uuid.UUID
	Side           order.Side
	State          order.State
	Type           order.Type
	Price          decimal.Decimal
	AvgPrice       decimal.Decimal
	Volume         decimal.Decimal
	OriginVolume   decimal.Decimal
	ExecutedVolume decimal.Decimal
	TradesCount    int
	Timestamp      int64
}

type OrderUpdate OrderMessage
type OrderCreate OrderMessage
type OrderReject OrderMessage
type OrderCancel OrderMessage

type UserTrade = tr.UserTrade

type OrderSnapshot struct {
	Market string
	Orders []OrderMessage
}

type OrderInfo struct {
	Orders []OrderMessage
}

type OrderTrades struct {
	Trades []UserTrade
}

type PublicTrade struct {
	Market    string
	ID        uint64
	Price     decimal.Decimal
	Amount    decimal.Decimal
	CreatedAt int64
	TakerSide order.Side
}

type OrderbookIncrement incremental.OrderbookIncrement

type OrderbookSnapshot incremental.OrderbookSnapshot

type BalanceUpdate struct {
	Snapshot []Balance
}

type Balance struct {
	Currency  string
	Available decimal.Decimal
	Locked    decimal.Decimal
}

type Response msg.Msg
