package bot

import (
	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/wcharczuk/go-chart/v2/drawing"

	"github.com/wcharczuk/go-chart/v2"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// showPriceHistoryHandler handles the inline button calls to the pricehistory button.
// It renders and sends a pricehistory chart to the user.
func showPriceHistoryHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	priceagent, getPriceagentErr := getPriceagentFromContext(ctx)
	if getPriceagentErr != nil {
		return getPriceagentErr
	}

	if priceagent.Entity.Type != geizhals.Product {
		_, _ = cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Pricehistory is only available for products at the moment!"})
		return nil
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "1M", CallbackData: fmt.Sprintf("m04_11_%d", priceagent.ID)},
				{Text: "🔘 3M", CallbackData: fmt.Sprintf("m04_12_%d", priceagent.ID)},
				{Text: "6M", CallbackData: fmt.Sprintf("m04_13_%d", priceagent.ID)},
				{Text: "12M", CallbackData: fmt.Sprintf("m04_14_%d", priceagent.ID)},
			},
			{
				{Text: "↩️ Zurück", CallbackData: fmt.Sprintf("m03_00_%d", priceagent.ID)},
			},
		},
	}

	_, _ = cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{})
	_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "upload_photo")
	history, err := geizhals.GetPriceHistory(priceagent.Entity)
	if err != nil {
		return fmt.Errorf("showPriceagentDetail: failed to download pricehistory: %w", err)
	}

	buffer := bytes.NewBuffer([]byte{})
	since := time.Now().AddDate(0, -3, 0)
	renderChart(priceagent, history, since, buffer)

	_, _ = bot.DeleteMessage(ctx.EffectiveChat.Id, cb.Message.MessageId)

	editedText := "Für welchen Zeitraum möchtest du die Preishistorie sehen?"
	_, sendErr := bot.SendPhoto(ctx.EffectiveUser.Id, buffer, &gotgbot.SendPhotoOpts{Caption: editedText, ReplyMarkup: markup})
	if sendErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to send photo: %w", sendErr)
	}
	return nil
}

// updatePriceHistoryGraphHandler handles the inline button calls to the date range buttons below the pricehistory chart.
// It renders an updated pricehistory chart and sends it to the user.
func updatePriceHistoryGraphHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	priceagent, getPriceagentErr := getPriceagentFromContext(ctx)
	if getPriceagentErr != nil {
		return getPriceagentErr
	}

	if priceagent.Entity.Type != geizhals.Product {
		_, _ = cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Pricehistory is only available for products at the moment!"})
		return nil
	}

	results := strings.Split(cb.Data, "_")
	if len(results) < 2 {
		return fmt.Errorf("updatePriceHistoryGraphHandler: invalid callback data: %s", cb.Data)
	}
	dateRange := results[1]

	dateRangeKeyboard, since := generateDateRangeKeyboard(priceagent, dateRange)

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			dateRangeKeyboard,
			{{Text: "↩️ Zurück", CallbackData: fmt.Sprintf("m03_00_%d", priceagent.ID)}},
		},
	}

	_, _ = cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{})
	history, err := geizhals.GetPriceHistory(priceagent.Entity)
	if err != nil {
		return fmt.Errorf("updatePriceHistoryGraphHandler: failed to download pricehistory: %w", err)
	}

	buffer := bytes.NewBuffer([]byte{})
	renderChart(priceagent, history, since, buffer)

	newPic := gotgbot.InputMediaPhoto{Media: buffer, Caption: "Für welchen Zeitraum möchtest du die Preishistorie sehen?"}
	_, sendErr := cb.Message.EditMedia(b, newPic, &gotgbot.EditMessageMediaOpts{ReplyMarkup: markup})
	if sendErr != nil {
		return fmt.Errorf("updatePriceHistoryGraphHandler: failed to send photo: %w", sendErr)
	}
	return nil
}

// generateDateRangeKeyboard generates the keyboard for the date range buttons below the pricehistory chart.
func generateDateRangeKeyboard(priceagent models.PriceAgent, dateRange string) ([]gotgbot.InlineKeyboardButton, time.Time) {
	dateRangeKeyboard := []gotgbot.InlineKeyboardButton{
		{Text: "1M", CallbackData: fmt.Sprintf("m04_11_%d", priceagent.ID)},
		{Text: "3M", CallbackData: fmt.Sprintf("m04_12_%d", priceagent.ID)},
		{Text: "6M", CallbackData: fmt.Sprintf("m04_13_%d", priceagent.ID)},
		{Text: "12M", CallbackData: fmt.Sprintf("m04_14_%d", priceagent.ID)},
	}

	var since time.Time
	switch dateRange {
	case "11":
		dateRangeKeyboard[0].Text = "🔘 1M"
		since = time.Now().AddDate(0, -1, 0)
	case "12":
		dateRangeKeyboard[1].Text = "🔘 3M"
		since = time.Now().AddDate(0, -3, 0)
	case "13":
		dateRangeKeyboard[2].Text = "🔘 6M"
		since = time.Now().AddDate(0, -6, 0)
	case "14":
		dateRangeKeyboard[3].Text = "🔘 12M"
		since = time.Now().AddDate(0, -12, 0)
	default:
		return generateDateRangeKeyboard(priceagent, "12")
	}
	return dateRangeKeyboard, since
}

// getPriceagentFromContext returns the priceagent from the callbackQuery data.
func getPriceagentFromContext(ctx *ext.Context) (models.PriceAgent, error) {
	cb := ctx.CallbackQuery
	priceagentID, parseErr := parseIDFromCallbackData(cb.Data, "m04_10_")
	if parseErr != nil {
		return models.PriceAgent{}, fmt.Errorf("showPriceagentDetail: failed to parse priceagentID from callback data: %w", parseErr)
	}

	priceagent, dbErr := database.GetPriceagentForUserByID(ctx.EffectiveUser.Id, priceagentID)
	if dbErr != nil {
		return models.PriceAgent{}, fmt.Errorf("showPriceagentDetail: failed to get priceagent from database: %w", dbErr)
	}
	return priceagent, nil
}

// renderChart renders a price history chart to the given writer.
func renderChart(priceagent models.PriceAgent, history geizhals.PriceHistory, since time.Time, w io.Writer) {
	darkFontColor := drawing.ColorFromHex("c2c2c2")
	fontColor := darkFontColor

	darkChartBackgroundColor := drawing.ColorFromHex("161b2b")
	chartBackgroundColor := darkChartBackgroundColor

	darkRegressionColor := drawing.ColorFromHex("e8a71a")
	regressionColor := darkRegressionColor

	darkMainSeriesColor := drawing.ColorFromHex("2569d1")
	mainSeriesColor := darkMainSeriesColor

	darkLegendBackgroundColor := drawing.ColorFromHex("2d364f")
	legendBackgroundColor := darkLegendBackgroundColor

	mainSeries := chart.TimeSeries{
		Name: priceagent.Name,
		Style: chart.Style{
			StrokeColor: mainSeriesColor,
			StrokeWidth: 2,
		},
	}

	maxPrice := 0.0
	minPrice := 9999999999999.0
	lastPrice := 0.0
	for _, entry := range history.Response {
		// Skip entries before given date
		if entry.Timestamp.Before(since) {
			// Still keep track of earlier prices for charts with gaps
			if entry.Valid {
				lastPrice = entry.Price
			}
			continue
		}

		if !entry.Valid {
			// When an entry isn't valid, use the last valid price
			if lastPrice == 0.0 {
				// If there is no last price, fully skip this entry - this can only happen for the first entries
				continue
			}

			mainSeries.XValues = append(mainSeries.XValues, entry.Timestamp)
			mainSeries.YValues = append(mainSeries.YValues, lastPrice)
			continue
		} else {
			mainSeries.XValues = append(mainSeries.XValues, entry.Timestamp)
			mainSeries.YValues = append(mainSeries.YValues, entry.Price)
			lastPrice = entry.Price
		}

		// Calculate current max/min price - only for valid entries
		if entry.Price > maxPrice {
			maxPrice = entry.Price
		}
		if entry.Price < minPrice {
			minPrice = entry.Price
		}
	}

	linRegSeries := &chart.LinearRegressionSeries{
		Name:        "Trend",
		InnerSeries: mainSeries,
		Style: chart.Style{
			StrokeColor:     regressionColor,
			StrokeWidth:     2,
			StrokeDashArray: []float64{5.0, 5.0},
		},
	}

	backgroundStyle := chart.Style{
		FillColor: chartBackgroundColor,
		FontColor: fontColor,
	}

	fontStyle := chart.Style{FontColor: fontColor}

	gridMajorStyle := chart.Style{
		Hidden:      false,
		StrokeColor: drawing.Color{R: 192, G: 192, B: 192, A: 100},
		StrokeWidth: 0.5,
	}

	gridMinorStyle := chart.Style{
		Hidden:      false,
		StrokeColor: drawing.Color{R: 192, G: 192, B: 192, A: 64},
		StrokeWidth: 0.5,
	}

	graph := chart.Chart{
		Width:      1280,
		Height:     720,
		Background: backgroundStyle,
		Canvas:     backgroundStyle,
		YAxis: chart.YAxis{
			Name: "Preis",
			Range: &chart.ContinuousRange{
				Min: minPrice - (maxPrice)*0.1,
				Max: maxPrice + (maxPrice)*0.1,
			},
			Style:          fontStyle,
			NameStyle:      fontStyle,
			GridMajorStyle: gridMajorStyle,
			GridMinorStyle: gridMinorStyle,
		},
		XAxis: chart.XAxis{
			Name:           "Datum",
			Style:          fontStyle,
			NameStyle:      fontStyle,
			GridMajorStyle: gridMajorStyle,
			GridMinorStyle: gridMinorStyle,
		},
		Series: []chart.Series{
			mainSeries,
			linRegSeries,
		},
	}
	graph.Elements = []chart.Renderable{chart.Legend(&graph, chart.Style{
		FillColor: legendBackgroundColor,
		FontColor: fontColor,
	})}

	renderErr := graph.Render(chart.PNG, w)
	if renderErr != nil {
		log.Println(renderErr)
	}
}
