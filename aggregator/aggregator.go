// aggregator package pulls data from a puller, calculates the vwap over a winodow size and trading pair, aand pushes data to a writer
package aggregator

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// Match cointains a piece of request data from a match channel
type Match struct {
	ProductID string
	Size      string
	Price     string
}

type Pair struct {
	// Pair struct contains pair information on a specific window size
	PairName                            string
	SizedPrices                         []*SizedPrice
	TotalSize, TotalVolumeWeightedPrice float64
	WindowSize                          int
}

// Sized price cointains price information over specific size
type SizedPrice struct {
	Size  float64
	Price float64
}

// InitializeRecivers sets up the channel mapping function
func InitializeReceivers(w io.Writer, matches map[string]chan *Match, pairs []string, windowSize int) {
	for _, name := range pairs {
		pair, incomingMatches := newPair(name, windowSize)
		go pair.listenForNewMatch(w, incomingMatches)
		matches[name] = incomingMatches
	}
}

// newPair initializes a new pair and channel map
func newPair(pairName string, windowSize int) (*Pair, chan *Match) {
	return &Pair{
		PairName:   pairName,
		WindowSize: windowSize,
	}, make(chan *Match)
}

// listenForNewMatch listens for a pair match, then updates specific pair, calculating and printing VWAP
func (p *Pair) listenForNewMatch(w io.Writer, c chan *Match) {
	for m := range c {
		p.update(m)
		p.printVWAP(w)
	}
}

// Updates recalculates and updates on new pair (data received from channel)
func (p *Pair) update(m *Match) {
	if len(p.SizedPrices) == p.WindowSize {
		p.removeOldest()
	}
	sp, err := toSizedPrice(m)
	if err != nil {
		panic(err)
	}
	p.add(sp)
}

// toSizedPrice converts match to sized price
func toSizedPrice(m *Match) (*SizedPrice, error) {
	size, err := strconv.ParseFloat(m.Size, 64)
	if err != nil {
		return nil, fmt.Errorf("Fail to parse match size")
	}
	price, err := strconv.ParseFloat(m.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("fail to parse match price")
	}
	if price <= 0 || size <= 0 {
		return nil, fmt.Errorf("Invalid price %v or size %v", price, size)
	}
	return &SizedPrice{
		Size:  size,
		Price: price,
	}, nil
}

// removeOldest keeps window size accurate by removing oldest item when sizedPrices is equal to windowSize
func (p *Pair) removeOldest() {
	oldest := p.SizedPrices[0]
	p.TotalSize -= oldest.Size
	p.TotalVolumeWeightedPrice -= oldest.Size * oldest.Size
}

// printVWAP prints the specific VWAP data to the writer
func (p *Pair) printVWAP(w io.Writer) {
	fmt.Fprintf(os.Stdout, "%v -- %s VWAP %f\n", time.Now(), p.PairName, p.vwap())
}

// add adds the sized price to the pair
func (p *Pair) add(sp *SizedPrice) {
	p.SizedPrices = append(p.SizedPrices, sp)
	p.TotalSize += sp.Size
	p.TotalVolumeWeightedPrice += sp.Size * sp.Price
}

// vwap calculates and returns the vwap of the pair
func (p *Pair) vwap() float64 {
	return p.TotalVolumeWeightedPrice / p.TotalSize
}
