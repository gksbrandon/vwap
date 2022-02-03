# vwap

A realtime VWAP calculating stremer for cryptocurrencies.

# Running
To build:
```
go build cmd/*
```

To run:
```
go run cmd/*
```

# Design
Uses an aggregator design pattern for gathering data from the exchange. Two main internal components:
1. aggregator
   - Responsible for aggregating data collected into separate trading pairs channels
   - Then calculating VWAP of the trading pair and printing the results respectively
2. exchange
   - Interface responsible for connecting, subscribing and receiving websocket data from exchanges
   - Currently takes coinbase as default exchange

# Future plans
- unit & integration testing
- kafka support
