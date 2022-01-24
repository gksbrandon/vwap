package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

const (
	origin = "http://localhost/"
	wsURL  = "wss://ws-feed.exchange.coinbase.com"
)

type Coinbase struct {
	Conn *websocket.Conn
}

type SubRequest struct {
	Type       RequestType `json:"type"`
	ProductIDs []string    `json:"product_ids"`
	Channels   []Channel   `json:"channels"`
}

type RequestType string

type SubResponse struct {
	Type      string    `json:"type"`
	Channels  []Channel `json:"channels"`
	Message   string    `json:"message,omitempty"`
	Size      string    `json:"size"`
	Price     string    `json:"price"`
	ProductID string    `json:"product_id"`
}

type Channel struct {
	Name       ChannelType
	ProductIDs []string
}
type ChannelType string

func (c Coinbase) Subscribe(ctx context.Context, ch chan Response, pairs []string) {
	conn, err := websocket.Dial(wsURL, "", origin)
	if err != nil {
		log.Fatal(fmt.Errorf("Websocket connected to websocket: %s, error: %v", wsURL, err))
	}
	log.Printf("Websocket connected to: %s", wsURL)

	subRequest := SubRequest{
		Type:       "subscribe",
		ProductIDs: pairs,
		Channels:   []Channel{{Name: "matches"}},
	}

	payload, err := json.Marshal(subRequest)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to marshal subscription: %w", err))
	}

	err = websocket.Message.Send(conn, payload)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to send subscription message: %w", err))
	}

	var subResponse SubResponse
	err = websocket.JSON.Receive(conn, &subResponse)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to write subscribe message: %w", err))
	}

	if subResponse.Type == "error" {
		log.Fatal(fmt.Errorf("Failed to subscribe, error message: %s", subResponse.Message))
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				err := conn.Close()
				if err != nil {
					log.Printf("Failed closing websocket connection: %s", err)
				}

			default:
				message := &Response{}
				err := websocket.JSON.Receive(conn, message)
				if err != nil {
					log.Printf("Failed receiving message: %s", err)
					break
				}
				ch <- *message
			}
		}
	}()
}
