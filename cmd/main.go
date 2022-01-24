package main

import (
	"context"
	"flag"
	"log"

	"github.com/gksbrandon/vwap/internal/aggregator"
	"github.com/gksbrandon/vwap/internal/exchange"
	"github.com/gksbrandon/vwap/internal/helper"
)

const (
	defaultTradingPairs = "BTC-USD, ETH-USD, ETH-BTC"
	defaultWindowSize   = 200
	defaultExchange     = "coinbase"
)

func main() {
	ctx := context.Background()

	// Command Line Flags
	tradingPairsFlag := flag.String("trading-pairs", defaultTradingPairs, "Trading pairs to subscribe to")
	windowSizeFlag := flag.Int("window-size", defaultWindowSize, "Window Size for VWAP calculation")
	exchangeFlag := flag.String("exchange", defaultExchange, "Exchange to stream trading data from")
	flag.Parse()

	// Validate & Split Trading Pairs Flag string ("ABC-XYZ, BCD-EFG") to Array ["ABC-XYZ", "BCD-EFG"]
	tradingPairs, err := helper.ValidateTradingPairs(tradingPairsFlag)
	if err != nil {
		log.Fatalf("Invalid Trading Pairs provided: %v", err)
	}

	// Intercepting shutdown signals -- for graceful shutdown with Ctrl-c
	go helper.InterceptShutdownSignals()

	// Initialize trading pair channel map and aggregator for each trading pair
	// Trading pair aggregator acts as receiver for respective trading pair
	// VWAP is then calculated and printed to screen
	ag := &aggregator.Aggregator{}
	ag.InitializeReceivers(tradingPairs, *windowSizeFlag)

	// Initialize exchange
	// Addional exchanges can be created from implementing the Client interface
	// And switching cases between them eg. ex.Client = &exchange.Binance
	ex := &exchange.Exchange{}
	switch *exchangeFlag {
	case "coinbase":
		ex.Client = &exchange.Coinbase{}
	default:
		log.Fatal("Please choose exchange from list: (coinbase)")
	}

	// Subscribe to exchange
	incomingMessages := make(chan exchange.Response)
	ex.Subscribe(ctx, incomingMessages, tradingPairs)

	// Receive data from subscribed channel in exchange, and send to respective aggregator
	for msg := range incomingMessages {
		m := &aggregator.Match{
			ProductID: msg.ProductID,
			Size:      msg.Size,
			Price:     msg.Price,
		}

		c, ok := ag.Matches[m.ProductID]
		if !ok {
			log.Println(m)
			log.Printf("no pair found for %s\n", m.ProductID)
			continue
		}
		c <- m
	}
}
