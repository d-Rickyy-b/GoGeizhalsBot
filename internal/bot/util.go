package bot

import (
	"GoGeizhalsBot/internal/geizhals"
	"fmt"
	"html"
)

// createLink generates a clickable html link given a display name and a url
func createLink(url, name string) string {
	return fmt.Sprintf("<a href=\"%s\">%s</a>", url, html.EscapeString(name))
}

// createPrice formats a given float to a formatted pricetag string
func createPrice(price float64) string {
	return fmt.Sprintf("%.2f €", price)
}

// bold encapsulates a string in a html <b> tag
func bold(text string) string {
	return fmt.Sprintf("<b>%s</b>", text)
}

// generateEntityKeyboard generates a gotgbot keyboard from a given list of entities
func generateEntityKeyboard(entities []geizhals.Entity) {
	// TODO: implement
}
