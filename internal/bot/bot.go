package bot

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/bot/models"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/bot/userstate"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/config"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/database"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/geizhals"
	"github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/prometheus"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
)

var bot *gotgbot.Bot

// startHandler is a message handler for the /start command.
func startHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Reset user's state to idle
	userID := ctx.EffectiveUser.Id
	userstate.UserStates[userID] = userstate.UserState{State: userstate.Idle}

	_, err := ctx.EffectiveMessage.Reply(bot, "Was möchtest du tun?", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
				{Text: "Neuer Preisagent", CallbackData: NewPriceAgentState},
				{Text: "Meine Preisagenten", CallbackData: ViewPriceAgentState},
			}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}

	return nil
}

// viewPriceagentsHandler is a callback handler that displays the first sub menu for the ViewPriceAgentState callback.
func viewPriceagentsHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	_, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{})
	if err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "📋 Wunschlisten", CallbackData: ShowWishlistPriceagentsState},
				{Text: "📦 Produkte", CallbackData: ShowProductPriceagentsState},
			},
			{
				{Text: "↩️ Zurück", CallbackData: MainMenuState},
			},
		},
	}
	_, _, err = cbq.Message.EditText(bot, "Welche Preisagenten möchtest du einsehen?", &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("viewPriceagents: failed to edit message text: %w", err)
	}

	return nil
}

// showWishlistPriceagents displays the menu with all wishlist priceagents for the ShowWishlistPriceagentsState callback
func showWishlistPriceagents(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	_, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{})
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

	markup := generateEntityKeyboard(priceagents, ShowPriceagentDetailState, 2)
	_, _, err = cbq.Message.EditText(bot, messageText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("showWishlist: failed to edit message text: %w", err)
	}

	return nil
}

// showProductPriceagents displays the menu with all product priceagents for the ShowProductPriceagentsState callback
func showProductPriceagents(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("showProductPriceagents: failed to answer callback query: %w", err)
	}

	productPriceagents, _ := database.GetProductPriceagentsForUser(ctx.EffectiveUser.Id)

	var messageText string
	if len(productPriceagents) == 0 {
		messageText = "Du hast noch keine Preisagenten für Produkte angelegt!"
	} else {
		messageText = "Das sind deine Preisagenten für deine Produkte:"
	}

	markup := generateEntityKeyboard(productPriceagents, ShowPriceagentDetailState, 2)
	_, _, err := cbq.Message.EditText(bot, messageText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("showProduct: failed to edit message text: %w", err)
	}

	return nil
}

// newPriceagentHandler is a callback handler for the NewPriceAgentState callback.
func newPriceagentHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	conf, confErr := config.GetConfig()
	if confErr != nil {
		return fmt.Errorf("newPriceagentHandler: failed to get config: %w", confErr)
	}

	cbq := ctx.Update.CallbackQuery

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	// check if user has capacities for a new priceagent
	if database.GetPriceAgentCountForUser(ctx.EffectiveUser.Id) >= conf.MaxPriceAgents {
		text := fmt.Sprintf("Du hast bereits die maximale Anzahl von %d Preisagenten angelegt. Bitte lösche einen Preisagenten, bevor du einen neuen anlegst.", conf.MaxPriceAgents)
		markup := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{Text: "Zu den Preisagenten", CallbackData: ViewPriceAgentState},
				},
			},
		}
		_, _, err := cbq.Message.EditText(bot, text, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
		if err != nil {
			return fmt.Errorf("newPriceagentHandler: failed to edit message text: %w", err)
		}

		return nil
	}
	_, _, err := cbq.Message.EditText(bot, "Bitte sende mir eine URL zu einem Produkt oder einer Wunschliste!", &gotgbot.EditMessageTextOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{}}})
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
func mainMenuHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
			{Text: "Neuer Preisagent", CallbackData: NewPriceAgentState},
			{Text: "Meine Preisagenten", CallbackData: ViewPriceAgentState},
		}},
	}

	_, _, err := cbq.Message.EditText(bot, "Was möchtest du tun?", &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("mainMenu: failed to edit message text: %w", err)
	}

	return nil
}

// showPriceagentDetail displays the menu for a single, specific price agent.
func showPriceagentDetail(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	menu, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to parse callback data: %w", parseErr)
	}

	var backCallbackData string

	switch {
	case priceagent.Entity.Type == geizhals.Wishlist:
		backCallbackData = ShowWishlistPriceagentsState
	case priceagent.Entity.Type == geizhals.Product:
		backCallbackData = ShowProductPriceagentsState
	default:
		backCallbackData = "invalidType"
	}

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("showPriceagentDetail: failed to answer callback query: %w", err)
	}

	notificationButtonText := fmt.Sprintf("⏰ %s", priceagent.NotificationSettings.String())

	linkName := createLink(priceagent.EntityURL(), priceagent.Entity.Name)
	price := priceagent.CurrentEntityPrice()
	editedText := fmt.Sprintf("%s kostet aktuell %s", linkName, bold(price.String()))
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: notificationButtonText, CallbackData: fmt.Sprintf("%s_%d", ChangePriceagentSettingsState, priceagent.ID)},
				{Text: "📊 Preisverlauf", CallbackData: fmt.Sprintf("%s_%d", ShowPriceHistoryState, priceagent.ID)},
			},
			{
				{Text: "❌ Löschen", CallbackData: fmt.Sprintf("%s_%d", DeletePriceagentConfirmState, priceagent.ID)},
				{Text: "↩️ Zurück", CallbackData: backCallbackData},
			},
		},
	}

	switch menu.SubMenu {
	case Menu0:
		_, _, err := cbq.Message.EditText(bot, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
		if err != nil {
			return fmt.Errorf("showPriceagentDetail: failed to edit message text: %w", err)
		}
	case Menu1:
		bot.DeleteMessage(ctx.EffectiveChat.Id, cbq.Message.GetMessageId(), nil)

		_, err := bot.SendMessage(ctx.EffectiveChat.Id, editedText, &gotgbot.SendMessageOpts{ReplyMarkup: markup, ParseMode: "HTML"})
		if err != nil {
			return fmt.Errorf("showPriceagentDetail: failed to send new message: %w", err)
		}
	case Menu2:
		_, err := bot.SendMessage(ctx.EffectiveChat.Id, editedText, &gotgbot.SendMessageOpts{ReplyMarkup: markup, ParseMode: "HTML"})
		if err != nil {
			return fmt.Errorf("showPriceagentDetail: failed to send new message: %w", err)
		}
	}

	return nil
}

// changePriceagentSettingsHandler handles the callbacks for the buttons to change the notification
// settings of a price agent.
func changePriceagentSettingsHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("changePriceagentSettingsHandler: failed to parse callback data: %w", parseErr)
	}

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("changePriceagentSettingsHandler: failed to answer callback query: %w", err)
	}

	linkName := createLink(priceagent.EntityURL(), priceagent.Entity.Name)
	price := priceagent.CurrentEntityPrice()
	editedText := fmt.Sprintf("%s\n\nWann möchtest du für %s alarmiert werden?\n\nAktuelle Einstellung: %s\nAktueller Preis: %s", bold("Benachrichtigungseinstellungen"), linkName, bold(priceagent.NotificationSettings.String()), bold(price.String()))
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "📉 Unter x€", CallbackData: fmt.Sprintf("%s_%d", SetNotificationBelowState, priceagent.ID)},
				{Text: "🔔 Immer", CallbackData: fmt.Sprintf("%s_%d", SetNotificationAlwaysState, priceagent.ID)},
			},
			{
				{Text: "↩️ Zurück", CallbackData: fmt.Sprintf("%s_%d", ShowPriceagentDetailState, priceagent.ID)},
			},
		},
	}
	_, _, err := cbq.Message.EditText(bot, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("showPriceagent: failed to edit message text: %w", err)
	}

	return nil
}

func deletePriceagentConfirmationHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		return fmt.Errorf("deletePriceagentConfirmationHandler: failed to parse callback data: %w", parseErr)
	}

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("deletePriceagentConfirmationHandler: failed to answer callback query: %w", err)
	}

	linkName := createLink(priceagent.EntityURL(), priceagent.Entity.Name)
	editedText := fmt.Sprintf("%s\n\nMöchtest du den Preisagenten für %s wirklich löschen?", bold("Löschen bestätigen"), linkName)
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "❌ Löschen", CallbackData: fmt.Sprintf("%s_%d", DeletePriceagentState, priceagent.ID)},
				{Text: "↩️ Zurück", CallbackData: fmt.Sprintf("%s_%d", ShowPriceagentDetailState, priceagent.ID)},
			},
		},
	}
	_, _, err := cbq.Message.EditText(bot, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("deletePriceagentConfirmationHandler: failed to edit message text: %w", err)
	}

	return nil
}

// deletePriceagentHandler handles all the inline "delete" buttons for priceagents
func deletePriceagentHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	// Get Priceagent from DB
	_, priceagent, parseErr := parseMenuPriceagent(ctx)
	if parseErr != nil {
		ctx.EffectiveMessage.Reply(bot, "Der Preisagent existiert nicht mehr, vielleicht wurde er schon gelöscht?", &gotgbot.SendMessageOpts{})
		return fmt.Errorf("deletePriceagentHandler: failed to parse callback data: %w", parseErr)
	}

	deleteErr := database.DeletePriceAgentForUser(priceagent)
	if deleteErr != nil {
		ctx.EffectiveMessage.Reply(bot, "Der Preisagent konnte nicht gelöscht werden!", &gotgbot.SendMessageOpts{})
		return fmt.Errorf("deletePriceagentHandler: failed to delete priceagent from database: %w", deleteErr)
	}

	editText := fmt.Sprintf("Preisagent für %s wurde gelöscht!", bold(createLink(priceagent.EntityURL(), priceagent.Entity.Name)))

	_, _, err := cbq.Message.EditText(bot, editText, &gotgbot.EditMessageTextOpts{ParseMode: "HTML", DisableWebPagePreview: true})
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
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(StopCancelState), stopHandlerCancel))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(StopConfirmState), stopHandlerConfirm))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(DeletePriceagentConfirmState), deletePriceagentConfirmationHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(DeletePriceagentState), deletePriceagentHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(ShowPriceHistoryState), showPriceHistoryHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(UpdateHistoryGraph1State), updatePriceHistoryGraphHandler))  // Graph 1M
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(UpdateHistoryGraph3State), updatePriceHistoryGraphHandler))  // Graph 3M
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(UpdateHistoryGraph6State), updatePriceHistoryGraphHandler))  // Graph 6M
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(UpdateHistoryGraph12State), updatePriceHistoryGraphHandler)) // Graph 12M
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(SetNotificationBelowState), setNotificationBelowHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(SetNotificationAlwaysState), setNotificationAlwaysHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(ChangePriceagentSettingsState), changePriceagentSettingsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix(ShowPriceagentDetailState), showPriceagentDetail))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(ShowWishlistPriceagentsState), showWishlistPriceagents))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(ShowProductPriceagentsState), showProductPriceagents))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(ViewPriceAgentState), viewPriceagentsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(NewPriceAgentState), newPriceagentHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(MainMenuState), mainMenuHandler))

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
	bot, createBotErr = gotgbot.NewBot(botConfig.BotToken, &gotgbot.BotOpts{})

	if createBotErr != nil {
		log.Println(botConfig)
		log.Fatalln("Something wrong:", createBotErr)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(_ *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
	})

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{
		ErrorLog: nil,
	})

	addMessageHandlers(dispatcher)
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
