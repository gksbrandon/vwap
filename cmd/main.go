package main

import (
	"context"
	"flag"
	"log"

	"github.com/gksbrandon/vwap/internal/aggregator"
	"github.com/gksbrandon/vwap/internal/exchange/coinbase"
	"github.com/gksbrandon/vwap/internal/helper"
)

const (
	defaultTradingPairs = "BTC-USD, ETH-USD, ETH-BTC"
	defaultWindowSize   = 200
	tradingPairsRegex   = `^([A-Z]{3}\-[A-Z]{3},)*([A-Z]{3}\-[A-Z]{3})$`
)

func main() {
	ctx := context.Background()

	// Command Line Flags
	tradingPairs := flag.String("trading-pairs", defaultTradingPairs, "Trading pairs to subscribe to")
	windowSize := flag.Int("window-size", defaultWindowSize, "Window Size for VWAP calculation")
	flag.Parse()

	// Validate Flags
	pairs := helper.ValidateTradingPairs(tradingPairs)

	// Intercepting shutdown signals.
	go helper.InterceptShutdownSignals()

	// Initialize trading pair channel map and aggregator for each pair
	tradingPairsChannelMap := make(map[string]chan *aggregator.Match)
	for _, name := range pairs {
		aggregator, incomingMatches := aggregator.NewPair(name, *windowSize)
		go aggregator.ListenForNewMatch(incomingMatches)
		tradingPairsChannelMap[name] = incomingMatches
	}

	// Subscribe to "matches" channel
	incomingMessages := make(chan coinbase.Response)
	client := coinbase.Client{}
	client.Subscribe(ctx, incomingMessages, pairs)

	// Receive data from exchange, send to respective aggregator
	for msg := range incomingMessages {
		m := &aggregator.Match{
			ProductID: msg.ProductID,
			Size:      msg.Size,
			Price:     msg.Price,
		}

		c, ok := tradingPairsChannelMap[m.ProductID]
		if !ok {
			log.Println(m)
			log.Printf("no pair found for %s\n", m.ProductID)
			continue
		}
		c <- m
	}
}
