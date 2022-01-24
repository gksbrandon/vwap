package coinbase

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

type Client struct {
	Conn *websocket.Conn
}

type Request struct {
	Type       RequestType `json:"type"`
	ProductIDs []string    `json:"product_ids"`
	Channels   []Channel   `json:"channels"`
}

type RequestType string

type Response struct {
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

func (c *Client) Connect() error {
	conn, err := websocket.Dial(wsURL, "", origin)
	if err != nil {
		return err
	}
	c.Conn = conn

	log.Printf("Websocket connected to: %s", wsURL)
	return nil
}

func (c *Client) Subscribe(ctx context.Context, ch chan Response, pairs []string) {
	err := c.Connect()
	if err != nil {
		log.Fatal(err)
	}

	subscriptionRequest := Request{
		Type:       "subscribe",
		ProductIDs: pairs,
		Channels:   []Channel{{Name: "matches"}},
	}

	payload, err := json.Marshal(subscriptionRequest)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to marshal subscription: %w", err))
	}

	err = websocket.Message.Send(c.Conn, payload)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to send subscription message: %w", err))
	}

	var subscriptionResponse Response
	err = websocket.JSON.Receive(c.Conn, &subscriptionResponse)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to write subscribe message: %w", err))
	}

	if subscriptionResponse.Type == "error" {
		log.Fatal(fmt.Errorf("Failed to subscribe, error message: %s", subscriptionResponse.Message))
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				err := c.Conn.Close()
				if err != nil {
					log.Printf("Failed closing websocket connection: %s", err)
				}

			default:
				message := &Response{}
				err := websocket.JSON.Receive(c.Conn, message)
				if err != nil {
					log.Printf("Failed receiving message: %s", err)
					break
				}
				ch <- *message
			}
		}
	}()
}
