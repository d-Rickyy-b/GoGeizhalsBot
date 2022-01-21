package bot

import (
	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/wcharczuk/go-chart/v2/drawing"

	"github.com/wcharczuk/go-chart/v2"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

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
		InnerSeries: mainSeries,
		Style: chart.Style{
			StrokeColor:     regressionColor,
			StrokeWidth:     2,
			StrokeDashArray: []float64{5.0, 5.0},
		},
	}

	graph := chart.Chart{
		Width:  1280,
		Height: 720,
		Background: chart.Style{
			FillColor: chartBackgroundColor,
			FontColor: fontColor,
		},
		Canvas: chart.Style{
			FillColor: chartBackgroundColor,
			FontColor: fontColor,
		},
		YAxis: chart.YAxis{
			Name: "Preis",
			Range: &chart.ContinuousRange{
				Min: minPrice - (maxPrice)*0.1,
				Max: maxPrice + (maxPrice)*0.1,
			},
			NameStyle: chart.Style{
				FontColor: fontColor,
			},
			Style: chart.Style{
				FontColor: fontColor,
			},
		},
		XAxis: chart.XAxis{
			Name: "Datum",
			Style: chart.Style{
				FontColor: fontColor,
			},
			NameStyle: chart.Style{
				FontColor: fontColor,
			},
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

// priceHistoryHandler handles the inline button calls to the pricehistory button.
// It renders and sends a pricehistory chart to the user.
func priceHistoryHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	priceagentID, parseErr := parseIDFromCallbackData(cb.Data, "m03_00_")
	if parseErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to parse priceagentID from callback data: %w", parseErr)
	}

	priceagent, dbErr := database.GetPriceagentForUserByID(ctx.EffectiveUser.Id, priceagentID)
	if dbErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to get priceagent from database: %w", dbErr)
	}

	if priceagent.Entity.Type != geizhals.Product {
		cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Pricehistory is only available for products at the moment!"})
		return nil
	}

	cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{})
	b.SendChatAction(ctx.EffectiveChat.Id, "upload_photo")
	history, err := geizhals.DownloadPriceHistory(priceagent.Entity)
	if err != nil {
		return fmt.Errorf("showPriceagentDetail: failed to download pricehistory: %w", err)
	}

	// TODO send photo
	buffer := bytes.NewBuffer([]byte{})
	renderChart(priceagent, history, time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), buffer)

	_, sendErr := bot.SendPhoto(ctx.EffectiveUser.Id, buffer, &gotgbot.SendPhotoOpts{})
	if sendErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to send photo: %w", sendErr)
	}
	return nil
}
