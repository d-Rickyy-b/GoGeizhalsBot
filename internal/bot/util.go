package bot

import (
	"GoGeizhalsBot/internal/bot/models"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// createLink generates a clickable html link given a display name and a url
func createLink(url, name string) string {
	name = strings.TrimSpace(name)
	return fmt.Sprintf("<a href=\"%s\">%s</a>", url, html.EscapeString(name))
}

// createPrice formats a given float to a formatted pricetag string
func createPrice(price float64) string {
	return fmt.Sprintf("%.2f €", price)
}

// bold encapsulates a string in a html <b> tag
func bold(text string) string {
	text = strings.TrimSpace(text)
	return fmt.Sprintf("<b>%s</b>", text)
}

// generateEntityKeyboard generates a gotgbot keyboard from a given list of entities
func generateEntityKeyboard(priceagents []models.PriceAgent, menuID string, numColumns int) gotgbot.InlineKeyboardMarkup {
	var keyboard [][]gotgbot.InlineKeyboardButton

	var row []gotgbot.InlineKeyboardButton //nolint:prealloc
	colCounter := 0
	for _, priceagent := range priceagents {
		row = append(row, gotgbot.InlineKeyboardButton{
			Text:         priceagent.Name,
			CallbackData: fmt.Sprintf("%s_%d", menuID, priceagent.ID),
		})
		colCounter++
		if colCounter%numColumns == 0 {
			keyboard = append(keyboard, row)
			row = []gotgbot.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	if len(priceagents) == 0 {
		keyboard = [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🆕 Neuer Preisagent", CallbackData: "m01_00"},
				{Text: "↩️ Zurück", CallbackData: "m01_01"},
			},
		}
	} else {
		// Add back button at the bottom row
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{{Text: "↩️ Zurück", CallbackData: "m01_01"}})
	}

	markup := gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard}
	return markup
}

func parseIDFromCallbackData(callbackData string, prefix string) (int64, error) {
	priceagentIDString := strings.TrimPrefix(callbackData, prefix)
	results := strings.Split(priceagentIDString, "_")

	// get last element from results
	priceagentIDString = results[len(results)-1]

	priceagentID, parseErr := strconv.Atoi(priceagentIDString)
	if parseErr != nil {
		return 0, parseErr
	}
	return int64(priceagentID), nil
}

func isAllowedLocation(location string) (allowed bool) {
	allowedLocations := []string{"de", "at", "uk", "pl"}
	for _, allowedLocation := range allowedLocations {
		if location == allowedLocation {
			return true
		}
	}
	return false
}
