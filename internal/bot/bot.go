package bot

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/bot/userstate"
	"GoGeizhalsBot/internal/config"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"GoGeizhalsBot/internal/prometheus"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
)

var bot *gotgbot.Bot

// startHandler is a message handler for the /start command.
func startHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	// Reset user's state to idle
	userID := ctx.EffectiveUser.Id
	userstate.UserStates[userID] = userstate.UserState{State: userstate.Idle}

	_, err := ctx.EffectiveMessage.Reply(b, "Was möchtest du tun?", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
				{Text: "Neuer Preisagent", CallbackData: "m01_00"},
				{Text: "Meine Preisagenten", CallbackData: "m01_01"},
			}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

// viewPriceagentsHandler is a callback handler that displays the first sub menu for the m01_01 callback.
func viewPriceagentsHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{})
	if err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "📋 Wunschlisten", CallbackData: "m02_00"},
				{Text: "📦 Produkte", CallbackData: "m02_01"},
			},
			{
				{Text: "↩️ Zurück", CallbackData: "m00_00"},
			},
		},
	}
	_, _, err = cb.Message.EditText(b, "Welche Preisagenten möchtest du einsehen?", &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("viewPriceagents: failed to edit message text: %w", err)
	}
	return nil
}

// showWishlistPriceagents displays the menu with all wishlist priceagents for the m02_00 callback
func showWishlistPriceagents(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{})
	if err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	priceagents, _ := database.GetWishlistPriceagentsForUser(ctx.EffectiveUser.Id)

	var messageText string
	if len(priceagents) == 0 {
		messageText = "Du hast noch keine Preisagenten für Wunschlisten angelegt!"
	} else {
		messageText = "Das sind deine Preisagenten für deine Wunschlisten:"
	}

	markup := generateEntityKeyboard(priceagents, "m03_00", 2)
	_, _, err = cb.Message.EditText(b, messageText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("showWishlist: failed to edit message text: %w", err)
	}
	return nil
}

// showProductPriceagents displays the menu with all product priceagents for the m02_01 callback
func showProductPriceagents(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("showProductPriceagents: failed to answer callback query: %w", err)
	}

	productPriceagents, _ := database.GetProductPriceagentsForUser(ctx.EffectiveUser.Id)

	var messageText string
	if len(productPriceagents) == 0 {
		messageText = "Du hast noch keine Preisagenten für Produkte angelegt!"
	} else {
		messageText = "Das sind deine Preisagenten für deine Produkte:"
	}

	markup := generateEntityKeyboard(productPriceagents, "m03_00", 2)
	_, _, err := cb.Message.EditText(b, messageText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("showProduct: failed to edit message text: %w", err)
	}
	return nil
}

// newPriceagentHandler is a callback handler for the m01_00 callback.
func newPriceagentHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("newPriceagentHandler")
	var maxNumberOfPriceagents int64 = 10
	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	// check if user has capacities for a new priceagent
	if database.GetPriceAgentCountForUser(ctx.EffectiveUser.Id) >= maxNumberOfPriceagents {
		_, _, err := cb.Message.EditText(b, "Du hast bereits 10 Preisagenten angelegt. Bitte lösche einen Preisagenten, bevor du einen neuen anlegst.", &gotgbot.EditMessageTextOpts{})
		if err != nil {
			return fmt.Errorf("newPriceagentHandler: failed to edit message text: %w", err)
		}
		return nil
	}
	_, _, err := cb.Message.EditText(b, "Bitte sende mir eine URL zu einem Produkt oder einer Wunschliste!", &gotgbot.EditMessageTextOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{}}})
	if err != nil {
		return fmt.Errorf("newPriceagent: failed to edit message text: %w", err)
	}

	// Set user's State
	userID := ctx.EffectiveUser.Id
	userstate.UserStates[userID] = userstate.UserState{State: userstate.CreatePriceagent}

	return nil
}

// mainMenuHandler handles all the back-button calls to the main menu.
// It's basically the same as the start handler, but without sending a new message.
func mainMenuHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
			{Text: "Neuer Preisagent", CallbackData: "m01_00"},
			{Text: "Meine Preisagenten", CallbackData: "m01_01"},
		}},
	}

	_, _, err := cb.Message.EditText(b, "Was möchtest du tun?", &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("mainMenu: failed to edit message text: %w", err)
	}
	return nil
}

// showPriceagentDetail displays the menu for a single, specific price agent.
func showPriceagentDetail(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	menu, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to parse callback data: %w", parseErr)
	}

	var backCallbackData string
	switch {
	case priceagent.Entity.Type == geizhals.Wishlist:
		backCallbackData = "m02_00"
	case priceagent.Entity.Type == geizhals.Product:
		backCallbackData = "m02_01"
	default:
		backCallbackData = "invalidType"
	}

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("showPriceagentDetail: failed to answer callback query: %w", err)
	}

	notificationButtonText := fmt.Sprintf("⏰ %s", priceagent.NotificationSettings.String())

	linkName := createLink(priceagent.EntityURL(), priceagent.Entity.Name)
	price := priceagent.CurrentEntityPrice()
	editedText := fmt.Sprintf("%s kostet aktuell %s", linkName, bold(price.String()))
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: notificationButtonText, CallbackData: fmt.Sprintf("m04_00_%d", priceagent.ID)},
				{Text: "📊 Preisverlauf", CallbackData: fmt.Sprintf("m05_00_%d", priceagent.ID)},
			},
			{
				{Text: "❌ Löschen", CallbackData: fmt.Sprintf("m04_98_%d", priceagent.ID)},
				{Text: "↩️ Zurück", CallbackData: backCallbackData},
			},
		},
	}

	switch menu.SubMenu {
	case "00":
		_, _, err := cb.Message.EditText(b, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
		if err != nil {
			return fmt.Errorf("showPriceagentDetail: failed to edit message text: %w", err)
		}
	case "01":
		bot.DeleteMessage(ctx.EffectiveChat.Id, cb.Message.MessageId, nil)

		_, err := b.SendMessage(ctx.EffectiveChat.Id, editedText, &gotgbot.SendMessageOpts{ReplyMarkup: markup, ParseMode: "HTML"})
		if err != nil {
			return fmt.Errorf("showPriceagentDetail: failed to send new message: %w", err)
		}
	case "02":
		_, err := b.SendMessage(ctx.EffectiveChat.Id, editedText, &gotgbot.SendMessageOpts{ReplyMarkup: markup, ParseMode: "HTML"})
		if err != nil {
			return fmt.Errorf("showPriceagentDetail: failed to send new message: %w", err)
		}
	}
	return nil
}

// changePriceagentSettingsHandler handles the callbacks for the buttons to change the notification
// settings of a price agent.
func changePriceagentSettingsHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("changePriceagentSettingsHandler: failed to parse callback data: %w", parseErr)
	}

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("changePriceagentSettingsHandler: failed to answer callback query: %w", err)
	}

	linkName := createLink(priceagent.EntityURL(), priceagent.Entity.Name)
	price := priceagent.CurrentEntityPrice()
	editedText := fmt.Sprintf("%s\n\nWann möchtest du für %s alarmiert werden?\n\nAktuelle Einstellung: %s\nAktueller Preis: %s", bold("Benachrichtigungseinstellungen"), linkName, bold(priceagent.NotificationSettings.String()), bold(price.String()))
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "📉 Unter x€", CallbackData: fmt.Sprintf("m04_02_%d", priceagent.ID)},
				{Text: "🔔 Immer", CallbackData: fmt.Sprintf("m04_01_%d", priceagent.ID)},
			},
			{
				{Text: "↩️ Zurück", CallbackData: fmt.Sprintf("m03_00_%d", priceagent.ID)},
			},
		},
	}
	_, _, err := cb.Message.EditText(b, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("showPriceagent: failed to edit message text: %w", err)
	}
	return nil
}

func deletePriceagentConfirmationHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("deletePriceagentConfirmationHandler: failed to parse callback data: %w", parseErr)
	}

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("deletePriceagentConfirmationHandler: failed to answer callback query: %w", err)
	}

	linkName := createLink(priceagent.EntityURL(), priceagent.Entity.Name)
	editedText := fmt.Sprintf("%s\n\nMöchtest du den Preisagenten für %s wirklich löschen?", bold("Löschen bestätigen"), linkName)
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "❌ Löschen", CallbackData: fmt.Sprintf("m04_99_%d", priceagent.ID)},
				{Text: "↩️ Zurück", CallbackData: fmt.Sprintf("m03_00_%d", priceagent.ID)},
			},
		},
	}
	_, _, err := cb.Message.EditText(b, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("deletePriceagentConfirmationHandler: failed to edit message text: %w", err)
	}
	return nil
}

// deletePriceagentHandler handles all the inline "delete" buttons for priceagents
func deletePriceagentHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	// Get Priceagent from DB
	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		ctx.EffectiveMessage.Reply(b, "Der Preisagent existiert nicht mehr, vielleicht wurde er schon gelöscht?", &gotgbot.SendMessageOpts{})
		return fmt.Errorf("deletePriceagentHandler: failed to parse callback data: %w", parseErr)
	}

	deleteErr := database.DeletePriceAgentForUser(priceagent)
	if deleteErr != nil {
		ctx.EffectiveMessage.Reply(b, "Der Preisagent konnte nicht gelöscht werden!", &gotgbot.SendMessageOpts{})
		return fmt.Errorf("deletePriceagentHandler: failed to delete priceagent from database: %w", deleteErr)
	}

	editText := fmt.Sprintf("Preisagent für %s wurde gelöscht!", bold(createLink(priceagent.EntityURL(), priceagent.Entity.Name)))

	_, _, err := cb.Message.EditText(b, editText, &gotgbot.EditMessageTextOpts{ParseMode: "HTML", DisableWebPagePreview: true})
	if err != nil {
		return fmt.Errorf("deletePriceagentHandler: failed to edit message text: %w", err)
	}
	return nil
}

func newUserHandler(_ *gotgbot.Bot, ctx *ext.Context) error {
	prometheus.TotalUserInteractions.Inc()
	// Create user in databse if they don't exist already
	if !ctx.EffectiveSender.IsUser() {
		return nil
	}

	user := models.User{
		ID:        ctx.EffectiveSender.User.Id,
		Username:  ctx.EffectiveSender.User.Username,
		FirstName: ctx.EffectiveSender.User.FirstName,
		LastName:  ctx.EffectiveSender.User.LastName,
		LangCode:  ctx.EffectiveSender.User.LanguageCode,
	}
	database.CreateUserWithCache(user)
	return nil
}

// setCommands sets all the available commands for the bot on Telegram
func setCommands() {
	_, setCommandErr := bot.SetMyCommands([]gotgbot.BotCommand{
		{Command: "start", Description: "Startmenü des Bots"},
		{Command: "stop", Description: "Löscht alle Daten und stoppt den Bot"},
		{Command: "help", Description: "Zeigt die Hilfe an"},
		{Command: "version", Description: "Zeigt die Version des Bots an"},
	}, &gotgbot.SetMyCommandsOpts{
		Scope:        gotgbot.BotCommandScopeDefault{},
		LanguageCode: "",
	})
	if setCommandErr != nil {
		log.Fatalln("Something wrong:", setCommandErr)
	}
}

// addMessageHandlers adds all the message handlers to the dispatcher. This tells our bot how to handle updates.
func addMessageHandlers(dispatcher *ext.Dispatcher) {
	// Text commands
	dispatcher.AddHandler(handlers.NewCommand("start", startHandler))
	dispatcher.AddHandler(handlers.NewCommand("stop", stopHandler))
	dispatcher.AddHandler(handlers.NewCommand("version", versionHandler))
	dispatcher.AddHandler(handlers.NewCommand("help", helpHandler))

	// Callback Queries (inline keyboards)
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m06_02"), stopHandlerCancel))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m06_01"), stopHandlerConfirm))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_98_"), deletePriceagentConfirmationHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_99_"), deletePriceagentHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m05_00_"), showPriceHistoryHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m05_01_"), updatePriceHistoryGraphHandler)) // Graph 1M
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m05_03_"), updatePriceHistoryGraphHandler)) // Graph 3M
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m05_06_"), updatePriceHistoryGraphHandler)) // Graph 6M
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m05_12_"), updatePriceHistoryGraphHandler)) // Graph 12M
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_02_"), setNotificationBelowHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_01_"), setNotificationAlwaysHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_00_"), changePriceagentSettingsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m03_"), showPriceagentDetail))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m02_00"), showWishlistPriceagents))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m02_01"), showProductPriceagents))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m01_01"), viewPriceagentsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m01_00"), newPriceagentHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m00_00"), mainMenuHandler))

	// Fallback handler for callback queries
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.All, fallbackCallbackHandler))

	// Unknown commands
	dispatcher.AddHandler(handlers.NewMessage(message.Command, fallbackCommandHandler))

	// Any kind of text
	dispatcher.AddHandler(handlers.NewMessage(message.Text, textHandler))

	// Store users if not already in database
	dispatcher.AddHandlerToGroup(handlers.NewCallback(callbackquery.All, newUserHandler), -1)
	dispatcher.AddHandlerToGroup(handlers.NewMessage(message.Text, newUserHandler), -1)
}

// Start is the main function to start the bot.
func Start(botConfig config.Config) {
	var createBotErr error
	bot, createBotErr = gotgbot.NewBot(botConfig.BotToken, &gotgbot.BotOpts{
		Client: http.Client{},
	})
	if createBotErr != nil {
		log.Println(botConfig)
		log.Fatalln("Something wrong:", createBotErr)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
	})

	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog:   nil,
		Dispatcher: dispatcher,
	})

	addMessageHandlers(updater.Dispatcher)
	setCommands()

	if botConfig.Webhook.Enabled {
		parsedURL, parseErr := url.Parse(botConfig.Webhook.URL)
		if parseErr != nil {
			log.Fatalln("Can't parse webhook url:", parseErr)
		}
		log.Printf("Starting webhook on '%s:%d%s'...\n", botConfig.Webhook.ListenIP, botConfig.Webhook.ListenPort, botConfig.Webhook.ListenPath)
		// TODO add support for custom certificates
		startErr := updater.StartWebhook(bot, parsedURL.Path, ext.WebhookOpts{
			ListenAddr: fmt.Sprintf("%s:%d", botConfig.Webhook.ListenIP, botConfig.Webhook.ListenPort),
		})
		if startErr != nil {
			panic("failed to start webhook: " + startErr.Error())
		}
		_, setWebhookErr := bot.SetWebhook(botConfig.Webhook.URL, &gotgbot.SetWebhookOpts{})
		if setWebhookErr != nil {
			panic("failed to set webhook: " + setWebhookErr.Error())
		}
	} else {
		log.Println("Start polling...")
		_, _ = bot.DeleteWebhook(nil)
		err := updater.StartPolling(bot, &ext.PollingOpts{DropPendingUpdates: false})
		if err != nil {
			panic("failed to start polling: " + err.Error())
		}
	}

	log.Printf("Bot has been started as @%s...\n", bot.User.Username)

	if botConfig.Prometheus.Enabled {
		// Periodically update the metrics from the database
		go func() {
			for {
				prometheus.TotalUniquePriceagentsValue = database.GetPriceAgentCount()
				prometheus.TotalUniqueUsersValue = database.GetUserCount()
				prometheus.TotalUniqueWishlistPriceagentsValue = database.GetPriceAgentWishlistCount()
				prometheus.TotalUniqueProductPriceagentsValue = database.GetPriceAgentProductCount()
				time.Sleep(time.Second * 60)
			}
		}()

		exportAddr := fmt.Sprintf("%s:%d", botConfig.Prometheus.ExportIP, botConfig.Prometheus.ExportPort)
		log.Printf("Starting prometheus exporter on %s...\n", exportAddr)
		err := prometheus.StartPrometheusExporter(exportAddr)
		if err != nil {
			panic("failed to start prometheus exporter: " + err.Error())
		}
	}

	updater.Idle()
}

func parseMenuPriceagent(ctx *ext.Context) (models.Menu, models.PriceAgent, error) {
	menu, parseMenuErr := models.NewMenu(ctx.CallbackQuery.Data)
	if parseMenuErr != nil {
		return models.Menu{}, models.PriceAgent{}, fmt.Errorf("invalid callback data: %s", ctx.CallbackQuery.Data)
	}

	priceAgent, dbErr := database.GetPriceagentForUserByID(ctx.EffectiveUser.Id, menu.PriceAgentID)
	if dbErr != nil {
		return models.Menu{}, models.PriceAgent{}, fmt.Errorf("invalid callback data: %s", ctx.CallbackQuery.Data)
	}

	return *menu, priceAgent, nil
}
