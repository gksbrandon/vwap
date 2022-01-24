package exchange

import "context"

type client interface {
	Subscribe(ctx context.Context, ch chan Response, pairs []string)
}

type Exchange struct {
	Client client
}

type Response struct {
	ProductID string `json:"product_id"`
	Size      string `json:"size"`
	Price     string `json:"price"`
}

func (e *Exchange) Subscribe(ctx context.Context, ch chan Response, pairs []string) {
	go e.Client.Subscribe(ctx, ch, pairs)
}
