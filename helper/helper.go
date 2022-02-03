// helper package contains helper functions such as validation functions
package helper

import (
	"fmt"
	"regexp"
	"strings"
)

// SetPairs validates the pairs against Coinbase's pairs format
func ValidateAndSplit(pairs string, regex string) ([]string, error) {
	// Remove all whitespaces
	s := strings.ReplaceAll(pairs, " ", "")

	// Regex to validate Coinbase pairs format
	r, _ := regexp.Compile(regex)
	if !r.MatchString(s) {
		return nil, fmt.Errorf("Invalid Trading Pairs provided: %s", pairs)
	}

	// Split by ',' and set pairs array
	return strings.Split(s, ","), nil
}
