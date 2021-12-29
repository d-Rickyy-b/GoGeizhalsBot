package bot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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

	return price, nil
}
