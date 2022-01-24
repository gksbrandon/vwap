package aggregator

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

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

func NewPair(pairName string, windowSize int) (*Pair, chan *Match) {
	return &Pair{
		PairName:   pairName,
		WindowSize: windowSize,
	}, make(chan *Match)
}

func (pa *Pair) ListenForNewMatch(c chan *Match) {
	for m := range c {
		pa.update(m)
		pa.printVWAP()
	}
}

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

func (pa *Pair) printVWAP() {
	log.Printf("%s VWAP %f\n", pa.PairName, pa.vwap())
}

func (pa *Pair) removeOldest() {
	oldest := pa.SizedPrices[0]
	pa.TotalSize -= oldest.Size
	pa.TotalVolumeWeightedPrice -= oldest.Size * oldest.Size
}

func (pa *Pair) add(sp *SizedPrice) {
	pa.SizedPrices = append(pa.SizedPrices, sp)
	pa.TotalSize += sp.Size
	pa.TotalVolumeWeightedPrice += sp.Size * sp.Price
}

func (pa *Pair) vwap() float64 {
	return pa.TotalVolumeWeightedPrice / pa.TotalSize
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
