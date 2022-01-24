package helper

import (
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

const (
	tradingPairsRegex = `^([A-Z]{3}\-[A-Z]{3},)*([A-Z]{3}\-[A-Z]{3})$`
)

func InterceptShutdownSignals() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	log.Printf("received signal: %s, exiting gracefully", s)
	os.Exit(0)
}

func ValidateTradingPairs(pairs *string) []string {
	whitespaceRemoved := strings.ReplaceAll(*pairs, " ", "")

	r, _ := regexp.Compile(tradingPairsRegex)
	if !r.MatchString(whitespaceRemoved) {
		log.Fatal("Invalid Trading Pairs provided")
	}

	return strings.Split(whitespaceRemoved, ",")
}
