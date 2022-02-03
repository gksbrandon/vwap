package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/gksbrandon/vwap/aggregator"
	"github.com/gksbrandon/vwap/exchange"
	"github.com/gksbrandon/vwap/exchange/coinbase"
	"github.com/gksbrandon/vwap/helper"
)

const (
	defaultTradingPairs = "BTC-USD,ETH-USD,ETH-BTC"
	defaultWindowSize   = 200
)

func main() {
	// Initialize Context
	ctx := context.Background()

	// Initialize Command Line Flags
	tradingPairsFlag := flag.String("trading-pairs", defaultTradingPairs, "Trading pairs to subscribe to eg. 'BTC-USD,ETH-USD,ETH-BTC'")
	windowSizeFlag := flag.Int("window-size", defaultWindowSize, "Window Size for VWAP calculation")
	flag.Parse()

	// Initialize Coinbase Connection
	// by choosing our connection, we can pass to our exchange as the client (allowing for different exchanges to be passed in the future)
	coin, err := coinbase.NewConnection()
	if err != nil {
		log.Fatalln(err)
	}

	// Validate and Split Trading Pairs Flag string ("ABC-XYZ, BCD-EFG") to Array ["ABC-XYZ", "BCD-EFG"]
	pairs, err := helper.ValidateAndSplit(*tradingPairsFlag, coin.Regex)
	if err != nil {
		log.Fatalln(err)
	}

	// Subscribe to Coinbase
	if err := coin.Subscribe(pairs); err != nil {
		log.Fatalln(err)
	}

	// Initialize Exchange
	// Additional exchanges can be created with their own Pull strategy and passed here
	ex := exchange.Exchange{
		Client: coin,
	}
	// Initialize Coinbase Message Receiver
	incoming := make(chan exchange.MatchesResponse)
	go ex.Pull(ctx, incoming)

	// Initialize Match Channel Map for each Trading Pair
	matches := make(map[string]chan *aggregator.Match)
	aggregator.InitializeReceivers(os.Stdout, matches, pairs, *windowSizeFlag)

	// Receive incoming data from subscribed channel in exchange, and push to respective aggregator
	for msg := range incoming {
		m := aggregator.Match{
			ProductID: msg.ProductID,
			Size:      msg.Size,
			Price:     msg.Price,
		}

		ch, ok := matches[m.ProductID]
		if !ok {
			log.Fatalf("no pair found for %s: %s\n", m.ProductID, m)
		}
		ch <- &m
	}
}
