package exchange

import (
	"context"
	"time"
)

type Puller interface {
	Pull(ctx context.Context, incoming chan MatchesResponse)
}

type Exchange struct {
	Client Puller
}

// matchesResponse is the "matches" data response received after subscribing to an exchange's channel
type MatchesResponse struct {
	Type         string    `json:"type"`
	TradeID      int       `json:"trade_id"`
	Sequence     int       `json:"sequence"`
	MakerOrderID string    `json:"maker_order_id"`
	TakerOrderID string    `json:"taker_order_id"`
	Time         time.Time `json:"time"`
	ProductID    string    `json:"product_id"`
	Size         string    `json:"size"`
	Price        string    `json:"price"`
	Side         string    `json:"side"`
}

// Pull calls the Client's implementation of Pull, allowing different exchanges to user different strategies for Pulling data
func (e *Exchange) Pull(ctx context.Context, incoming chan MatchesResponse) {
	e.Client.Pull(ctx, incoming)

}
