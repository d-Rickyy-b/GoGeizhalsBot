package bot

import (
	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/bot/userstate"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"errors"
	"fmt"
	"log"

	"github.com/mattn/go-sqlite3"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// textHandler handles all the text messages. It checks if the message contains a valid URL and if so, it creates a new pricehandler
func textHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveUser.Id
	log.Printf("User sent '%s'\n", ctx.EffectiveMessage.Text)

	var (
		state userstate.UserState
		ok    bool
	)

	if state, ok = userstate.UserStates[userID]; !ok {
		state = userstate.UserState{
			State:      userstate.Idle,
			Priceagent: models.PriceAgent{},
		}
		userstate.UserStates[userID] = state
	}
	ctx.Data["state"] = state

	switch state.State {
	case userstate.CreatePriceagent:
		return textNewPriceagentHandler(b, ctx)
	case userstate.SetNotification:
		return textChangeNotificationSettingsHandler(b, ctx)
	}

	// Parse link and request price
	return nil
}

// textChangeNotificationSettingsHandler handles the text message when the user wants to change the notification settings of a price agent
func textChangeNotificationSettingsHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveUser.Id
	var outOfRangePostfix string

	price, parseErr := parsePrice(ctx.EffectiveMessage.Text)
	if parseErr != nil {
		log.Printf("parsePrice: %s\n", parseErr)
		if errors.Is(parseErr, ErrOutOfRange) {
			outOfRangePostfix = fmt.Sprintf("Der Preis liegt außerhalb des gültigen Bereichs und wurde daher auf %.2f € gesetzt.", price)
		} else {
			ctx.EffectiveMessage.Reply(b, "Bitte sende mir einen Preis in der Form: '3,99' oder '3.99'!", &gotgbot.SendMessageOpts{})
			return nil
		}
	}

	newNotifSettings := models.NotificationSettings{
		NotifyBelow: true,
		BelowPrice:  price,
	}

	state := ctx.Data["state"].(userstate.UserState)

	dbErr := database.UpdateNotificationSettings(userID, state.Priceagent.ID, newNotifSettings)
	if dbErr != nil {
		log.Printf("UpdateNotificationSettings: %s\n", dbErr)
		ctx.EffectiveMessage.Reply(b, "Es ist ein Fehler beim Speichern der Einstellungen aufgetreten!", &gotgbot.SendMessageOpts{})
		return dbErr
	}

	markup := gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "Zum Preisagenten!", CallbackData: fmt.Sprintf("m03_00_%d", state.Priceagent.ID)},
		},
	}}

	messageText := fmt.Sprintf("Preisagent wurde bearbeitet! %s", outOfRangePostfix)
	b.SendMessage(ctx.EffectiveChat.Id, messageText, &gotgbot.SendMessageOpts{ReplyMarkup: markup})
	return nil
}

// textNewPriceagentHandler handles text messages that contain a link to a geizhals product or wishlist
func textNewPriceagentHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("User in CreatePriceagent state!")

	entity, downloadErr := geizhals.DownloadEntity(ctx.EffectiveMessage.Text)
	if downloadErr != nil {
		ctx.EffectiveMessage.Reply(b, "Bitte sende eine valide Geizhals URL!", &gotgbot.SendMessageOpts{})
		log.Printf("textNewPriceagentHandler: %s\n", downloadErr)
		return nil
	}

	newPriceagent := models.PriceAgent{
		//ID:     entity.ID,
		Name:   entity.Name,
		UserID: ctx.EffectiveUser.Id,
		Entity: entity,
		NotificationSettings: models.NotificationSettings{
			NotifyAlways: true,
		},
	}

	createErr := database.CreatePriceAgentForUser(&newPriceagent)
	if createErr != nil {
		// TODO check if user already has price agent for this entity
		if errors.Is(createErr, sqlite3.ErrConstraintUnique) {
			log.Printf("Priceagent already exists: %s\n", createErr)
			ctx.EffectiveMessage.Reply(b, "Ein Preisagent für diese/s Produkt/Wunschliste existiert bereits!", &gotgbot.SendMessageOpts{})
			return createErr
		}
		log.Printf("CreatePriceAgentForUser: %s\n", createErr)
		ctx.EffectiveMessage.Reply(b, "Es ist ein Fehler aufgetreten!", &gotgbot.SendMessageOpts{})
		return createErr
	}

	markup := gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
		{
			{Text: "Zum Preisalarm!", CallbackData: fmt.Sprintf("m03_00_%d", newPriceagent.ID)},
		},
	}}
	b.SendMessage(ctx.EffectiveChat.Id, "Preisagent wurde erstellt!", &gotgbot.SendMessageOpts{ReplyMarkup: markup})
	// TODO send user to priceagent overview
	return nil
}
