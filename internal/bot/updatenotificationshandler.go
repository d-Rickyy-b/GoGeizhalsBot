package bot

import (
	"fmt"
	"log"

	"github.com/d-Rickyy-b/gogeizhalsbot/internal/bot/models"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/bot/userstate"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/database"
	"github.com/d-Rickyy-b/gogeizhalsbot/internal/geizhals"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// setNotificationBelowHandler handles callback queries for the option to set notifications to appear
// when the price drops below a certain price
func setNotificationBelowHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("setNotificationBelowHandler: failed to parse callback data: %w", parseErr)
	}

	userID := ctx.EffectiveUser.Id
	userstate.UserStates[userID] = userstate.UserState{State: userstate.SetNotification, Priceagent: priceagent}

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("setNotificationBelowHandler: failed to answer callback query: %w", err)
	}

	entityPrice := priceagent.CurrentEntityPrice()
	editedText := fmt.Sprintf("Ab welchem Preis möchtest du für %s alarmiert werden?\nAktueller Preis: %s", createLink(priceagent.EntityURL(), priceagent.Name), bold(entityPrice.String()))
	_, _, err := cb.Message.EditText(b, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{}, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("setNotificationBelowHandler: failed to edit message text: %w", err)
	}

	return nil
}

// setNotificationAlwaysHandler handles callback queries for the option to set notifications to appear
// at any change of the price
func setNotificationAlwaysHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("setNotificationAlwaysHandler: failed to parse callback data: %w", parseErr)
	}

	newNotifSettings := models.NotificationSettings{
		NotifyAlways: true,
	}

	dbUpdateErr := database.UpdateNotificationSettings(ctx.EffectiveUser.Id, priceagent.ID, newNotifSettings)
	if dbUpdateErr != nil {
		log.Printf("UpdateNotificationSettings: %s\n", dbUpdateErr)
		ctx.EffectiveMessage.Reply(bot, "Es ist ein Fehler aufgetreten!", &gotgbot.SendMessageOpts{})

		return dbUpdateErr
	}

	// Notify user about their decision, then go back to the priceagent detail overview
	text := "Du wirst ab sofort für jede Preisänderung benachrichtigt!"

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{Text: text}); err != nil {
		return fmt.Errorf("setNotificationAlwaysHandler: failed to answer callback query: %w", err)
	}

	var backCallbackData string
	switch priceagent.Entity.Type {
	case geizhals.Wishlist:
		backCallbackData = "m02_00"
	case geizhals.Product:
		backCallbackData = "m02_01"
	default:
		backCallbackData = "invalidType"
	}

	linkName := createLink(priceagent.EntityURL(), priceagent.Entity.Name)
	price := priceagent.CurrentEntityPrice()
	editedText := fmt.Sprintf("%s kostet aktuell %s", linkName, bold(price.String()))
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "⏰ Benachrichtigung", CallbackData: fmt.Sprintf("m04_00_%d", priceagent.ID)},
				{Text: "📊 Preisverlauf", CallbackData: fmt.Sprintf("m05_00_%d", priceagent.ID)},
			},
			{
				{Text: "❌ Löschen", CallbackData: fmt.Sprintf("m04_98_%d", priceagent.ID)},
				{Text: "↩️ Zurück", CallbackData: backCallbackData},
			},
		},
	}
	_, _, err := cb.Message.EditText(b, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("showPriceagent: failed to edit message text: %w", err)
	}

	return nil
}
