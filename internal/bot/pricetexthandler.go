package bot

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var ErrOutOfRange = errors.New("price is out of range")

// parsePrice tries to parse a price as float from a given string.
func parsePrice(messageText string) (float64, error) {
	priceRegex := regexp.MustCompile(`^\s*(\d+(?:[,.]\d+)?)\s*€?$`)
	priceRegexMatch := priceRegex.FindStringSubmatch(messageText)

	if len(priceRegexMatch) == 0 {
		return 0, fmt.Errorf("could not parse price from message text: %s", messageText)
	}

	priceString := priceRegexMatch[1]
	priceString = strings.ReplaceAll(priceString, ",", ".")

	price, parseError := strconv.ParseFloat(priceString, 64)
	if parseError != nil {
		return 0, fmt.Errorf("could not parse price from message text: %s", messageText)
	}

	// check if price is in range
	upperBound := 1000000.00
	lowerBound := 0.01

	if price < lowerBound {
		return lowerBound, ErrOutOfRange
	} else if price > upperBound {
		return upperBound, ErrOutOfRange
	}

	return price, nil
}
