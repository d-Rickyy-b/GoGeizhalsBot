package bot

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"time"

	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/bot/models"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/database"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/geizhals"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/prometheus"

	"github.com/d-Rickyy-b/go-chart-x/v2/pkg/chart"
	"github.com/d-Rickyy-b/go-chart-x/v2/pkg/drawing"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// showPriceHistoryHandler handles the inline button calls to the pricehistory button.
// It renders and sends a pricehistory chart to the user.
func showPriceHistoryHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("showPriceHistoryHandler: failed to parse callback data: %w", parseErr)
	}

	isDarkmode := database.GetDarkmode(ctx.EffectiveUser.Id)
	dateRangeKeyboard, since := generateDateRangeKeyboard(priceagent, "03", isDarkmode)
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			dateRangeKeyboard,
			{{Text: "↩️ Zurück", CallbackData: fmt.Sprintf("%s_%d", ShowPriceagentDetailState, priceagent.ID)}},
		},
	}

	_, _ = bot.SendChatAction(ctx.EffectiveChat.Id, "upload_photo", nil)
	history, err := geizhals.GetPriceHistory(priceagent.Entity, priceagent.Location)
	if err != nil {
		return fmt.Errorf("showPriceagentDetail: failed to download pricehistory: %w", err)
	}

	if len(history.Response) == 0 {
		log.Println("showPriceagentDetail: pricehistory is empty")
		_, _ = bot.AnswerCallbackQuery(cbq.Id, &gotgbot.AnswerCallbackQueryOpts{Text: "Preisverlaufdaten konnten nicht aktualisiert werden."})
		return nil
	}

	buffer := bytes.NewBuffer([]byte{})
	renderChart(priceagent, history, since, buffer, isDarkmode)

	_, _ = bot.DeleteMessage(ctx.EffectiveChat.Id, cbq.Message.GetMessageId(), nil)

	editedText := fmt.Sprintf("%s\nFür welchen Zeitraum möchtest du die Preishistorie sehen?", bold(createLink(priceagent.EntityURL(), priceagent.Name)))
	_, sendErr := bot.SendPhoto(ctx.EffectiveUser.Id, buffer, &gotgbot.SendPhotoOpts{Caption: editedText, ReplyMarkup: markup, ParseMode: "HTML"})
	if sendErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to send photo: %w", sendErr)
	}

	return nil
}

// updatePriceHistoryGraphHandler handles the inline button calls to the date range buttons below the pricehistory chart.
// It renders an updated pricehistory chart and sends it to the user.
func updatePriceHistoryGraphHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	menu, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("updatePriceHistoryGraphHandler: failed to parse callback data: %w", parseErr)
	}

	darkMode := database.GetDarkmode(ctx.EffectiveUser.Id)

	if menu.Extra != "" {
		changeDarkmodeTo := menu.Extra
		switch changeDarkmodeTo {
		case "0":
			darkMode = false
		default:
			darkMode = true
		}
	}

	database.UpdateDarkMode(ctx.EffectiveUser.Id, darkMode)

	dateRange := menu.SubMenu
	dateRangeKeyboard, since := generateDateRangeKeyboard(priceagent, dateRange, darkMode)

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			dateRangeKeyboard,
			{{Text: "↩️ Zurück", CallbackData: fmt.Sprintf("%s_%d", ShowPriceagentDetailState, priceagent.ID)}},
		},
	}

	history, err := geizhals.GetPriceHistory(priceagent.Entity, priceagent.Location)
	if err != nil {
		_, _ = cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Preisverlaufdaten konnten nicht aktualisiert werden."})
		return fmt.Errorf("updatePriceHistoryGraphHandler: failed to download pricehistory: %w", err)
	}

	if len(history.Response) == 0 {
		_, _ = cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: "Preisverlaufdaten konnten nicht aktualisiert werden."})
		log.Println("updatePriceHistoryGraphHandler: pricehistory is empty")

		return nil
	}

	_, _ = cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{})
	buffer := bytes.NewBuffer([]byte{})
	renderChart(priceagent, history, since, buffer, darkMode)

	caption := fmt.Sprintf("%s\nFür welchen Zeitraum möchtest du die Preishistorie sehen?", bold(createLink(priceagent.EntityURL(), priceagent.Name)))
	newPic := gotgbot.InputMediaPhoto{Media: buffer, Caption: caption, ParseMode: "HTML"}
	_, _, sendErr := cbq.Message.EditMedia(bot, newPic, &gotgbot.EditMessageMediaOpts{ReplyMarkup: markup})
	if sendErr != nil {
		return fmt.Errorf("updatePriceHistoryGraphHandler: failed to send photo: %w", sendErr)
	}

	return nil
}

// generateDateRangeKeyboard generates the keyboard for the date range buttons below the pricehistory chart.
func generateDateRangeKeyboard(priceagent models.PriceAgent, dateRange string, isDarkmode bool) ([]gotgbot.InlineKeyboardButton, time.Time) {
	themeButton := "🌑"
	switchTheme := 1

	if isDarkmode {
		themeButton = "🌕"
		switchTheme = 0
	}

	dateRangeKeyboard := []gotgbot.InlineKeyboardButton{
		{Text: "1M", CallbackData: fmt.Sprintf("%s_%d", UpdateHistoryGraph1State, priceagent.ID)},
		{Text: "3M", CallbackData: fmt.Sprintf("%s_%d", UpdateHistoryGraph3State, priceagent.ID)},
		{Text: "6M", CallbackData: fmt.Sprintf("%s_%d", UpdateHistoryGraph6State, priceagent.ID)},
		{Text: "12M", CallbackData: fmt.Sprintf("%s_%d", UpdateHistoryGraph12State, priceagent.ID)},
		{Text: themeButton, CallbackData: fmt.Sprintf("m05_%s_%d_%d", dateRange, priceagent.ID, switchTheme)},
	}

	var since time.Time

	switch dateRange {
	case "01":
		dateRangeKeyboard[0].Text = "🔘 1M"
		since = time.Now().AddDate(0, -1, 0)
	case "03":
		dateRangeKeyboard[1].Text = "🔘 3M"
		since = time.Now().AddDate(0, -3, 0)
	case "06":
		dateRangeKeyboard[2].Text = "🔘 6M"
		since = time.Now().AddDate(0, -6, 0)
	case "12":
		dateRangeKeyboard[3].Text = "🔘 12M"
		since = time.Now().AddDate(0, -12, 0)
	default:
		return generateDateRangeKeyboard(priceagent, "03", isDarkmode)
	}

	return dateRangeKeyboard, since
}

// renderChart renders a price history chart to the given writer.
func renderChart(priceagent models.PriceAgent, history geizhals.PriceHistory, since time.Time, w io.Writer, darkmode bool) {
	prometheus.GraphsRendered.Inc()
	var fontColor, chartBackgroundColor, regressionColor, mainSeriesColor, legendBackgroundColor, gridMajorStrokeColor, gridMinorStrokeColor drawing.Color

	fontColor = chart.DefaultTextColor
	chartBackgroundColor = chart.DefaultBackgroundColor
	regressionColor = drawing.ColorFromHex("e8a71a")
	mainSeriesColor = drawing.ColorFromHex("2569d1")
	legendBackgroundColor = chart.DefaultBackgroundColor
	gridMajorStrokeColor = drawing.Color{R: 192, G: 192, B: 192, A: 100}
	gridMinorStrokeColor = drawing.Color{R: 192, G: 192, B: 192, A: 64}

	if darkmode {
		fontColor = drawing.ColorFromHex("c2c2c2")
		chartBackgroundColor = drawing.ColorFromHex("161b2b")
		legendBackgroundColor = drawing.ColorFromHex("2d364f")
	}

	mainSeries := chart.TimeSeries{
		Name: fmt.Sprintf("Preis (%s)", priceagent.CurrentEntityPrice().Currency.String()),
		Style: chart.Style{
			StrokeColor: mainSeriesColor,
			StrokeWidth: 3,
		},
	}

	maxPrice := 0.0
	minPrice := math.MaxFloat64
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
		maxPrice = math.Max(maxPrice, entry.Price)
		minPrice = math.Min(minPrice, entry.Price)
	}

	linRegSeries := &chart.LinearRegressionSeries{
		Name:        "Trend",
		InnerSeries: mainSeries,
		Style: chart.Style{
			StrokeColor:     regressionColor,
			StrokeWidth:     3,
			StrokeDashArray: []float64{5.0, 5.0},
		},
	}

	backgroundStyle := chart.Style{
		FillColor: chartBackgroundColor,
		FontColor: fontColor,
	}

	fontStyle := chart.Style{FontColor: fontColor, FontSize: 12}

	gridMajorStyle := chart.Style{
		Hidden:      false,
		StrokeColor: gridMajorStrokeColor,
		StrokeWidth: 0.5,
	}

	gridMinorStyle := chart.Style{
		Hidden:      false,
		StrokeColor: gridMinorStrokeColor,
		StrokeWidth: 0.5,
	}

	graph := chart.Chart{
		Title: priceagent.Name,
		TitleStyle: chart.Style{
			FontColor: fontColor,
			FontSize:  12,
		},
		Width:      1280,
		Height:     720,
		Background: backgroundStyle,
		Canvas:     backgroundStyle,
		YAxis: chart.YAxis{
			Name: fmt.Sprintf("Preis (%s)", priceagent.CurrentEntityPrice().Currency.String()),
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
	graph.Elements = []chart.Renderable{chart.LegendLineLeft(&graph, chart.Style{
		FillColor: legendBackgroundColor,
		FontColor: fontStyle.FontColor,
		FontSize:  fontStyle.FontSize,
	})}

	renderErr := graph.Render(chart.PNG, w)
	if renderErr != nil {
		log.Println(renderErr)
	}
}
