package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gksbrandon/vwap/exchange"
	"golang.org/x/net/websocket"
)

const (
	origin     = "http://localhost/"
	URL        = "wss://ws-feed.exchange.coinbase.com"
	pairsRegex = `^([A-Z]{3}\-[A-Z]{3},)*([A-Z]{3}\-[A-Z]{3})$`
)

// Coinbase struct keeps state of coinbase exchange server
type Coinbase struct {
	URL   string
	Conn  *websocket.Conn
	Regex string
}

// NewCoinbase works as the Factory Function Coinbase exchange
func NewConnection() (*Coinbase, error) {
	conn, err := websocket.Dial(URL, "", origin)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to websocket: %s, error: %v", URL, err)
	}
	log.Printf("Websocket connected to: %s", URL)
	return &Coinbase{
		URL:   URL,
		Conn:  conn,
		Regex: pairsRegex,
	}, nil
}

// subscriptionRequest is the request made to the exchange to subscribe to a channel and product
type subscriptionRequest struct {
	Type       string    `json:"type"`
	ProductIDs []string  `json:"product_ids"`
	Channels   []channel `json:"channels"`
}

// channel is the channel object passed in a subscriptionRequest
type channel struct {
	Name       string
	ProductIDs []string
}

// subscriptionResponse is the response received from a subscription request to an exchange
type subscriptionResponse struct {
	Type      string    `json:"type"`
	Channels  []channel `json:"channels"`
	Message   string    `json:"message,omitempty"`
	Size      string    `json:"size"`
	Price     string    `json:"price"`
	ProductID string    `json:"product_id"`
}

// Subscribe to Coinbase exchange server, indicating which channels and products to receive
func (c *Coinbase) Subscribe(pairs []string) error {
	// Send Subscription Request
	subRequest := subscriptionRequest{
		Type:       "subscribe",
		ProductIDs: pairs,
		Channels: []channel{
			{Name: "matches", ProductIDs: pairs},
		},
	}
	payload, err := json.Marshal(subRequest)
	if err != nil {
		return fmt.Errorf("Failed to marshal subscription: %w", err)
	}
	err = websocket.Message.Send(c.Conn, payload)
	if err != nil {
		return fmt.Errorf("Failed to send subscription message: %w", err)
	}

	// Receive Subscription Response
	var subResponse subscriptionResponse
	err = websocket.JSON.Receive(c.Conn, &subResponse)
	if err != nil {
		return fmt.Errorf("Failed to write subscribe message: %w", err)
	}
	if subResponse.Type == "error" {
		return fmt.Errorf("Failed to subscribe, error message: %s", subResponse.Message)
	}

	log.Println("Subscribed to matches channel")
	return nil
}

// Pull allows the receiver channel passed in to receive JSON messages from the exchange, based on the context lifetime
func (c *Coinbase) Pull(ctx context.Context, incoming chan exchange.MatchesResponse) {
	for {
		select {
		case <-ctx.Done():
			err := c.Conn.Close()
			if err != nil {
				log.Printf("Failed closing websocket connection: %s", err)
			}
		default:
			message := exchange.MatchesResponse{}
			err := websocket.JSON.Receive(c.Conn, &message)
			if err != nil {
				log.Printf("Failed receiving message: %s", err)
			}
			if message.Type == "subscriptions" {
				continue
			}
			incoming <- message
		}
	}
}
