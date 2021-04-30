package opendax

import (
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
)

// Opendax is the overarching type across this package
type Opendax struct {
	exchange.Base
}

type Client struct {
	conn *websocket.Conn

	key    string
	secret string

	done chan struct{}
	msgs chan *msg.Msg
}

const (
	opendaxAPIURL     = ""
	opendaxAPIVersion = ""

	// Public endpoints

	// Authenticated endpoints
)

// Start implementing public and private exchange API funcs below

func New(key, secret string) *Client {
	return &Client{
		key:    key,
		secret: secret,
		done:   make(chan struct{}),
		msgs:   make(chan *msg.Msg),
	}
}

func (c *Client) apikeyHeaders(header http.Header) http.Header {
	kid := c.key
	secret := c.secret
	nonce := fmt.Sprintf("%d", time.Now().Unix()*1000)
	data := nonce + kid

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	sha := hex.EncodeToString(h.Sum(nil))

	header.Add("X-Auth-Apikey", kid)
	header.Add("X-Auth-Nonce", nonce)
	header.Add("X-Auth-Signature", sha)

	return header
}

func (c *Client) Connect(url string) error {
	body := bytes.NewBuffer(nil)
	conn, resp, err := websocket.DefaultDialer.Dial(url, c.apikeyHeaders(http.Header{}))
	if resp.Body != nil {
		defer resp.Body.Close()
		io.Copy(body, resp.Body)
	}
	if err != nil {
		return fmt.Errorf("%w: %s", err, body.String())
	}
	resp.Body.Close()
	c.conn = conn

	return nil
}

func (c *Client) Listen() <-chan *msg.Msg {
	go func() {
		defer func() {
			close(c.done)
			close(c.msgs)
		}()

		for {
			_, m, err := c.conn.ReadMessage()
			if err != nil {
				log.Debug("websocket", "error on read message", err.Error())
				return
			}

			parsed, err := msg.Parse(m)
			if err != nil {
				log.Debug("websocket", "error on parse message", err.Error())
				continue
			}

			c.msgs <- parsed
		}
	}()

	return c.msgs
}

func (c *Client) Shutdown() error {
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}

	select {
	case <-c.done:
	case <-time.After(time.Second):
	}

	return c.conn.Close()
}

func (c *Client) Send(msg []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *Client) SendMsg(msg *msg.Msg) error {
	b, err := msg.Encode()
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.TextMessage, b)
}

func SubscribePrivate(reqID uint64, topics ...interface{}) *msg.Msg {
	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodSubscribe,
		Args: []interface{}{
			"private",
			topics,
		},
	}
}

func SubscribePublic(reqID uint64, topics ...interface{}) *msg.Msg {
	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodSubscribe,
		Args: []interface{}{
			"public",
			topics,
		},
	}
}

func ListOpenOrders(reqID uint64, market string) *msg.Msg {
	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodListOrders,
		Args:   []interface{}{market},
	}
}

func GetOrders(reqID uint64, list ...uuid.UUID) *msg.Msg {
	args := make([]interface{}, len(list))
	for i, uuid := range list {
		args[i] = uuid
	}

	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodGetOrders,
		Args:   args,
	}
}

func GetOrderTrades(reqID uint64, uuid uuid.UUID) *msg.Msg {
	return &msg.Msg{
		ReqID:  reqID,
		Type:   msg.Request,
		Method: protocol.MethodGetOrderTrades,
		Args:   []interface{}{uuid},
	}
}

func orderType(t order.Type) string {
	switch t {
	case order.Limit:
		return protocol.OrderTypeLimit
	case order.Market:
		return protocol.OrderTypeMarket
	case order.PostOnly:
		return protocol.OrderTypePostOnly
	case order.FOK:
		return protocol.OrderTypeFillOrKill
	default:
		return "???"
	}
}

func OrderSide(s order.Side) string {
	switch s {
	case order.Buy:
		return protocol.OrderSideBuy
	case order.Sell:
		return protocol.OrderSideSell
	default:
		return "?"
	}
}

func orderState(s order.State) string {
	switch s {
	case order.StateCancel:
		return protocol.OrderStateCancel
	case order.StateDone:
		return protocol.OrderStateDone
	case order.StateWait:
		return protocol.OrderStateWait
	case order.StatePending:
		return protocol.OrderStatePending
	case order.StateReject:
		return protocol.OrderStateReject
	default:
		return "?"
	}
}

// [1,42,"createOrder",["btcusd", "M", "S", "0.250000", "9120.00"]]
func OrderCreateReq(reqID uint64, o *order.Model) *msg.Msg {
	return msg.NewRequest(
		reqID,
		protocol.MethodCreateOrder,
		o.Market,
		orderType(o.Type),
		OrderSide(o.Side),
		o.Volume.String(),
		o.Price.String(),
	)
}

func OrderCancelReq(reqID uint64, orderUUID string) *msg.Msg {
	return msg.NewRequest(reqID, protocol.MethodCancelOrder, "uuid", orderUUID)
}



func Parse(m *msg.Msg) (interface{}, error) {
	switch m.Type {
	case msg.Response:
		return recognizeResponse(m)

	case msg.EventPrivate:
		return recognizeEventPrivate(m)

	case msg.EventPublic:
		return recognizeEventPublic(m)

	default:
		return m, nil
	}
}

func recognizeResponse(m *msg.Msg) (interface{}, error) {
	switch m.Method {
	case protocol.MethodListOrders:
		return orderSnapshot(m.Args)

	case protocol.MethodGetOrders:
		return orderInfo(m.Args)

	case protocol.MethodGetOrderTrades:
		return orderTrades(m.Args)

	default:
		return (*Response)(m), nil
	}
}

func orderTrades(args []interface{}) (*OrderTrades, error) {
	it := msg.NewArgIterator(args)
	snapshot := OrderTrades{}

	for {
		trArgs, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break
		}
		if err != nil {
			return &snapshot, err
		}

		tr, err := trade(trArgs)
		if err != nil {
			return &snapshot, err
		}

		snapshot.Trades = append(snapshot.Trades, *tr)
	}

	return &snapshot, nil
}

func orderInfo(args []interface{}) (*OrderInfo, error) {
	it := msg.NewArgIterator(args)
	snapshot := OrderInfo{}

	for {
		ordArgs, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break
		}
		if err != nil {
			return &snapshot, err
		}

		ord, err := orderMessage(ordArgs)
		if err != nil {
			return &snapshot, err
		}

		snapshot.Orders = append(snapshot.Orders, *ord)
	}

	return &snapshot, nil
}

func orderSnapshot(args []interface{}) (*OrderSnapshot, error) {
	it := msg.NewArgIterator(args)
	snapshot := OrderSnapshot{}

	market, err := it.NextString()
	if err != nil {
		return &snapshot, err
	}

	snapshot.Market = market

	for {
		ordArgs, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break
		}
		if err != nil {
			return &snapshot, err
		}

		ord, err := orderMessage(ordArgs)
		if err != nil {
			return &snapshot, err
		}

		snapshot.Orders = append(snapshot.Orders, *ord)
	}

	return &snapshot, nil
}

func publicTrade(args []interface{}) (interface{}, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	id, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	price, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	amount, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	ts, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	tSide, err := it.NextString()
	if err != nil {
		return nil, err
	}

	takerSide, err := recognizeSide(tSide)
	if err != nil {
		return nil, err
	}

	return &PublicTrade{
		Market:    market,
		ID:        id,
		Price:     price,
		Amount:    amount,
		CreatedAt: int64(ts),
		TakerSide: takerSide,
	}, nil
}

func priceLevel(args []interface{}) (incremental.PriceLevel, error) {
	it := msg.NewArgIterator(args)

	price, err := it.NextDecimal()
	if err != nil {
		return incremental.PriceLevel{}, err
	}

	amountString, err := it.NextString()
	if err != nil {
		return incremental.PriceLevel{}, err
	}

	if amountString == "" {
		return incremental.PriceLevel{Price: price, Amount: decimal.Zero}, nil
	}

	amount, err := decimal.NewFromString(amountString)
	if err != nil {
		return incremental.PriceLevel{}, err
	}

	return incremental.PriceLevel{
		Price:  price,
		Amount: amount,
	}, nil
}

// ["btcusd",1,"asks",["9120","0.25"]]
func orderbookIncrement(args []interface{}) (*OrderbookIncrement, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	seq, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	s, err := it.NextString()
	if err != nil {
		return nil, err
	}

	var side order.Side
	switch s {
	case "asks":
		side = order.Sell
	case "bids":
		side = order.Buy
	default:
		return nil, errors.New("invalid increment side " + s)
	}

	slice, err := it.NextSlice()
	if err != nil {
		return nil, err
	}

	pl, err := priceLevel(slice)
	if err != nil {
		return nil, err
	}

	return &OrderbookIncrement{
		Market:     market,
		Sequence:   seq,
		Side:       side,
		PriceLevel: pl,
	}, nil
}

// ["btcusd",1,[["9120","0.25"]],[]]
func orderbookSnapshot(args []interface{}) (*OrderbookSnapshot, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	seq, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	slice, err := it.NextSlice()
	if err != nil {
		return nil, err
	}

	sells, err := getPriceLevels(slice)
	if err != nil {
		return nil, err
	}

	slice, err = it.NextSlice()
	if err != nil {
		return nil, err
	}

	buys, err := getPriceLevels(slice)
	if err != nil {
		return nil, err
	}

	return &OrderbookSnapshot{
		Market:   market,
		Sequence: seq,
		Buys:     buys,
		Sells:    sells,
	}, nil
}

func getPriceLevels(args []interface{}) ([]incremental.PriceLevel, error) {
	it := msg.NewArgIterator(args)
	res := make([]incremental.PriceLevel, 0)

	for {
		slice, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break
		}

		if err != nil {
			return nil, err
		}

		pl, err := priceLevel(slice)
		if err != nil {
			return nil, err
		}

		res = append(res, pl)
	}

	return res, nil
}

func recognizeEventPublic(m *msg.Msg) (interface{}, error) {
	switch m.Method {
	case "trade":
		return publicTrade(m.Args)
	case "obi":
		return orderbookIncrement(m.Args)
	case "obs":
		return orderbookSnapshot(m.Args)
	default:
		return m, errors.New("unexpected message")
	}
}

func recognizeEventPrivate(m *msg.Msg) (interface{}, error) {
	switch m.Method {
	case protocol.EventOrderCreate:
		parsed, err := orderMessage(m.Args)
		return (*OrderCreate)(parsed), err

	case protocol.EventOrderCancel:
		parsed, err := orderMessage(m.Args)
		return (*OrderCancel)(parsed), err

	case protocol.EventOrderUpdate:
		parsed, err := orderMessage(m.Args)
		return (*OrderUpdate)(parsed), err

	case protocol.EventOrderReject:
		parsed, err := orderMessage(m.Args)
		return (*OrderReject)(parsed), err

	case protocol.EventTrade:
		return trade(m.Args)

	case protocol.EventBalanceUpdate:
		return balanceUpdate(m.Args)

	default:
		return m, errors.New("unexpected message")
	}
}

func balanceUpdate(args []interface{}) (*BalanceUpdate, error) {
	it := msg.NewArgIterator(args)

	balances := make([]Balance, 0, len(args))

loop:
	for {
		bu, err := it.NextSlice()
		if err == msg.ErrIterationDone {
			break loop
		}
		if err != nil {
			return nil, err
		}

		buIt := msg.NewArgIterator(bu)
		cur, err := buIt.NextString()
		if err != nil {
			return nil, err
		}

		avail, err := buIt.NextDecimal()
		if err != nil {
			return nil, err
		}

		locked, err := buIt.NextDecimal()
		if err != nil {
			return nil, err
		}

		balances = append(balances, Balance{Currency: cur, Available: avail, Locked: locked})
	}

	return &BalanceUpdate{balances}, nil
}

func trade(args []interface{}) (*UserTrade, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	id, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	price, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	amount, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	total, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	orderID, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	orderUUID, err := it.NextUUID()
	if err != nil {
		return nil, err
	}

	side, err := it.NextString()
	if err != nil {
		return nil, err
	}

	ordSide, err := recognizeSide(side)
	if err != nil {
		return nil, err
	}

	tSide, err := it.NextString()
	if err != nil {
		return nil, err
	}

	takerSide, err := recognizeSide(tSide)
	if err != nil {
		return nil, err
	}

	fee, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	feeCur, err := it.NextString()
	if err != nil {
		return nil, err
	}

	ts, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	return &UserTrade{
		Market:      market,
		ID:          id,
		Price:       price,
		Amount:      amount,
		Total:       total,
		OrderID:     orderID,
		OrderUUID:   orderUUID,
		OrderSide:   ordSide,
		TakerSide:   takerSide,
		Fee:         fee,
		FeeCurrency: feeCur,
		CreatedAt:   int64(ts),
	}, nil
}

func MessageFromUserTrade(tr *UserTrade) []interface{} {
	return []interface{}{
		tr.Market,
		tr.ID,
		tr.Price,
		tr.Amount,
		tr.Total,
		tr.OrderID,
		tr.OrderUUID,
		OrderSide(tr.OrderSide),
		OrderSide(tr.TakerSide),
		tr.Fee,
		tr.FeeCurrency,
		tr.CreatedAt,
	}
}

func MessageFromOrder(o *order.Model) []interface{} {
	return []interface{}{
		o.Market,
		o.ID,
		o.UUID,
		OrderSide(o.Side),
		orderState(o.State),
		orderType(o.Type),
		o.Price,
		o.AveragePrice(),
		o.Volume,
		o.OriginVolume,
		o.OriginVolume.Sub(o.Volume),
		o.TradesCount,
		o.CreatedAt,
	}
}

func recognizeSide(side string) (order.Side, error) {
	switch side {
	case protocol.OrderSideSell:
		return order.Sell, nil
	case protocol.OrderSideBuy:
		return order.Buy, nil
	default:
		return "?", errors.New("order side invalid: " + side)
	}
}

func orderMessage(args []interface{}) (*OrderMessage, error) {
	it := msg.NewArgIterator(args)

	market, err := it.NextString()
	if err != nil {
		return nil, err
	}

	id, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	uuidParsed, err := it.NextUUID()
	if err != nil {
		return nil, err
	}

	side, err := it.NextString()
	if err != nil {
		return nil, err
	}

	ordSide, err := recognizeSide(side)
	if err != nil {
		return nil, err
	}

	var ordState order.State
	state, err := it.NextString()
	if err != nil {
		return nil, err
	}

	switch state {
	case protocol.OrderStatePending:
		ordState = order.StatePending
	case protocol.OrderStateWait:
		ordState = order.StateWait
	case protocol.OrderStateDone:
		ordState = order.StateDone
	case protocol.OrderStateReject:
		ordState = order.StateReject
	case protocol.OrderStateCancel:
		ordState = order.StateCancel
	default:
		return nil, errors.New("unexpected order state ")
	}

	var ordType order.Type
	typ, err := it.NextString()
	if err != nil {
		return nil, err
	}

	switch typ {
	case protocol.OrderTypeMarket:
		ordType = order.Market
	case protocol.OrderTypeLimit:
		ordType = order.Limit
	case protocol.OrderTypePostOnly:
		ordType = order.PostOnly
	default:
		return nil, errors.New("unexpected order type")
	}

	price, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	avgPrice, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	volume, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	originVolume, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	executed, err := it.NextDecimal()
	if err != nil {
		return nil, err
	}

	trades, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	ts, err := it.NextUint64()
	if err != nil {
		return nil, err
	}

	return &OrderMessage{
		Market:         market,
		ID:             id,
		UUID:           uuidParsed,
		Side:           ordSide,
		State:          ordState,
		Type:           ordType,
		Price:          price,
		AvgPrice:       avgPrice,
		Volume:         volume,
		OriginVolume:   originVolume,
		ExecutedVolume: executed,
		TradesCount:    int(trades),
		Timestamp:      int64(ts),
	}, nil
