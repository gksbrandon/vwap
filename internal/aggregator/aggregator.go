package aggregator

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

type Aggregator struct {
	Matches map[string]chan *Match
}

type Pair struct {
	PairName                            string
	SizedPrices                         []*SizedPrice
	TotalSize, TotalVolumeWeightedPrice float64
	WindowSize                          int
}

type SizedPrice struct {
	Size  float64
	Price float64
}

type Match struct {
	ProductID string `json:"product_id"`
	Size      string `json:"size"`
	Price     string `json:"price"`
}

// Creates a mapping for each trading pair to a Receiver Matches channel
func (a *Aggregator) InitializeReceivers(tradingPairs []string, windowSize int) {
	a.Matches = make(map[string]chan *Match)
	for _, name := range tradingPairs {
		pair, incomingMatches := newPair(name, windowSize)
		go pair.listenForNewMatch(incomingMatches)
		a.Matches[name] = incomingMatches
	}
}

// Creates a new pair and channel map
func newPair(pairName string, windowSize int) (*Pair, chan *Match) {
	return &Pair{
		PairName:   pairName,
		WindowSize: windowSize,
	}, make(chan *Match)
}

// Channel listening for a pair match, then updating, calculating and printing VWAP
func (p *Pair) listenForNewMatch(c chan *Match) {
	for m := range c {
		p.update(m)
		p.printVWAP()
	}
}

// Updates new pair match data received from channel
func (pa *Pair) update(m *Match) {
	if len(pa.SizedPrices) == pa.WindowSize {
		pa.removeOldest()
	}
	sp, err := toSizedPrice(m)
	if err != nil {
		panic(err)
	}
	pa.add(sp)
}

func toSizedPrice(m *Match) (*SizedPrice, error) {
	size, err := strconv.ParseFloat(m.Size, 64)
	if err != nil {
		return nil, errors.New("Fail to parse match size")
	}
	price, err := strconv.ParseFloat(m.Price, 64)
	if err != nil {
		return nil, errors.New("fail to parse match price")
	}
	if price <= 0 || size <= 0 {
		return nil, fmt.Errorf("Invalid price %v or size %v", price, size)
	}
	return &SizedPrice{
		Size:  size,
		Price: price,
	}, nil
}

func (pa *Pair) removeOldest() {
	oldest := pa.SizedPrices[0]
	pa.TotalSize -= oldest.Size
	pa.TotalVolumeWeightedPrice -= oldest.Size * oldest.Size
}

func (pa *Pair) printVWAP() {
	log.Printf("%s VWAP %f\n", pa.PairName, pa.vwap())
}

func (pa *Pair) add(sp *SizedPrice) {
	pa.SizedPrices = append(pa.SizedPrices, sp)
	pa.TotalSize += sp.Size
	pa.TotalVolumeWeightedPrice += sp.Size * sp.Price
}

func (pa *Pair) vwap() float64 {
	return pa.TotalVolumeWeightedPrice / pa.TotalSize
}
