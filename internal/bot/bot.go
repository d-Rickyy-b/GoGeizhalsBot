package bot

import (
	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/bot/userstate"
	"GoGeizhalsBot/internal/config"
	"GoGeizhalsBot/internal/database"
	"GoGeizhalsBot/internal/geizhals"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
)

var bot *gotgbot.Bot

// updateEntityPrices fetches the current price of all entities and updates the database
func updateEntityPrices() {
	allEntities, fetchEntitiesErr := database.GetAllEntities()
	if fetchEntitiesErr != nil {
		log.Println("Error fetching entites:", fetchEntitiesErr)
		return
	}

	// Iterate over all price agents.
	// For each price agent, update prices and store updated prices in the entity in the database.
	// Also update price history with the new prices.
	for _, entity := range allEntities {
		log.Println("Updating prices for:", entity.Name)

		// If there are two price agents with the same entity, we currently fetch it twice
		updatedEntity, updateErr := geizhals.UpdateEntity(entity)
		if updateErr != nil {
			log.Println("Error updating entity:", updateErr)
			continue
		}

		if updatedEntity.Price == entity.Price {
			log.Println("Entity price has not changed, skipping update")
			continue
		}

		entityPrice := models.HistoricPrice{
			Price:    updatedEntity.Price,
			EntityID: updatedEntity.ID,
			Entity:   updatedEntity,
		}

		database.AddHistoricPrice(entityPrice)
		database.UpdateEntity(updatedEntity)

		// TODO notify users
		// fetch all priceagents for this entity
		priceAgents, fetchPriceAgentsErr := database.GetPriceAgentsForEntity(updatedEntity.ID)
		if fetchPriceAgentsErr != nil {
			log.Println("Error fetching price agents for entity:", fetchPriceAgentsErr)
			continue
		}

		for _, priceAgent := range priceAgents {
			notifyUsers(priceAgent, entity, updatedEntity)
		}
	}
}

// notifyUsers sends a notification to the users of the price agent if the settings allow it
func notifyUsers(priceAgent models.PriceAgent, oldEntity, updatedEntity geizhals.Entity) {
	settings := priceAgent.NotificationSettings
	user := priceAgent.User
	diff := updatedEntity.Price - oldEntity.Price

	var change string
	if updatedEntity.Price > oldEntity.Price {
		change = fmt.Sprintf("📈 %s teurer", bold(createPrice(diff)))
	} else {
		change = fmt.Sprintf("📉 %s günstiger", bold(createPrice(diff)))
	}

	var notificationText string
	entityLink := createLink(updatedEntity.URL, updatedEntity.Name)
	entityPrice := bold(createPrice(updatedEntity.Price))
	if settings.NotifyAlways {
		notificationText = fmt.Sprintf("Der Preis von %s hat sich geändert: %s\n\n%s", entityLink, entityPrice, change)
	} else if settings.NotifyBelow && updatedEntity.Price < settings.BelowPrice {
		notificationText = fmt.Sprintf("Der Preis von %s hat sich geändert: %s\n\n%s", entityLink, entityPrice, change)
	} else if settings.NotifyAbove && updatedEntity.Price > settings.AbovePrice {
		notificationText = "Hi, preis über Grenze!"
	} else if settings.NotifyPriceDrop && updatedEntity.Price < oldEntity.Price {
		notificationText = "Hi, preis gefallen!"
	} else if settings.NotifyPriceRise && updatedEntity.Price > oldEntity.Price {
		notificationText = "Hi, preis gestiegen!"
	} else {
		log.Println("Price changes don't match the notification settings for user")
		return
	}

	// TODO implement message queueing to avoid hitting telegram api limits (30 msgs/sec)
	_, sendErr := bot.SendMessage(user.ID, notificationText, &gotgbot.SendMessageOpts{ParseMode: "HTML", DisableWebPagePreview: true})
	if sendErr != nil {
		log.Println("Error sending message:", sendErr)
		return
	}
}

// UpdatePricesJob is a job that updates prices of all price agents at a given interval.
func UpdatePricesJob(updateFrequency time.Duration) {
	// Initialize lastCheck time to now-2*updateFrequency to ensure that the first update is done immediately.
	lastCheck := time.Now().Add(-2 * updateFrequency)

	// Align method execution at certain intervals - e.g. every 5 minutes at :05, :10, :15, etc. similar to cron.
	delta := time.Now().Unix() % int64(updateFrequency.Seconds())
	initialDelay := updateFrequency - (time.Second * time.Duration(delta))
	log.Println("Initial Delay:", initialDelay)
	time.Sleep(initialDelay)

	for {
		// calculate difference between now and lastCheck,
		passedTime := time.Since(lastCheck)
		// if it's < updateFrequency, sleep for the remaining time
		if passedTime < updateFrequency {
			sleepDuration := updateFrequency - passedTime
			log.Println("Sleeping for:", sleepDuration)
			time.Sleep(sleepDuration)
		}

		// check for price updates
		updateEntityPrices()
		lastCheck = time.Now()
	}
}

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
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
			{Text: "📋 Wunschlisten", CallbackData: "m02_00"},
			{Text: "📦 Produkte", CallbackData: "m02_01"},
		},
			{
				{Text: "↩️ Zurück", CallbackData: "m00_00"},
			},
		},
	}
	_, err = cb.Message.EditText(b, "Welche Preisagenten möchtest du einsehen?", &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("viewPriceagents: failed to edit message text: %w", err)
	}
	return nil
}

// showWishlistPriceagents displays the menu with all wishlist priceagents for the m02_00 callback
func showWishlistPriceagents(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("showWishlist")
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

	markup := generateEntityKeyboard(priceagents, "m03_00_", 2)
	_, err = cb.Message.EditText(b, messageText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("showWishlist: failed to edit message text: %w", err)
	}
	return nil
}

// showProductPriceagents displays the menu with all product priceagents for the m02_01 callback
func showProductPriceagents(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("showProduct")
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

	markup := generateEntityKeyboard(productPriceagents, "m03_00_", 2)
	_, err := cb.Message.EditText(b, messageText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
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
		_, err := cb.Message.EditText(b, "Du hast bereits 10 Preisagenten angelegt. Bitte lösche einen Preisagenten, bevor du einen neuen anlegst.", &gotgbot.EditMessageTextOpts{})
		if err != nil {
			return fmt.Errorf("newPriceagentHandler: failed to edit message text: %w", err)
		}
		return nil
	}
	_, err := cb.Message.EditText(b, "Bitte sende mir eine URL zu einem Produkt oder einer Wunschliste!", &gotgbot.EditMessageTextOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{}}})
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

	_, err := cb.Message.EditText(b, "Was möchtest du tun?", &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("mainMenu: failed to edit message text: %w", err)
	}
	return nil
}

// showPriceagentDetail displays the menu for a single, specific price agent.
func showPriceagentDetail(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	priceagentID, parseErr := parseIDFromCallbackData(cb.Data, "m03_00_")
	if parseErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to parse priceagentID from callback data: %w", parseErr)
	}

	priceagent, dbErr := database.GetPriceagentForUserByID(ctx.EffectiveUser.Id, priceagentID)
	if dbErr != nil {
		return fmt.Errorf("showPriceagentDetail: failed to get priceagent from database: %w", dbErr)
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

	linkName := createLink(priceagent.Entity.URL, priceagent.Entity.Name)
	editedText := fmt.Sprintf("%s kostet aktuell %s", linkName, bold(createPrice(priceagent.Entity.Price)))
	markup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "⏰ Benachrichtigung", CallbackData: fmt.Sprintf("m04_00_%d", priceagent.ID)},
				{Text: "📊 Preisverlauf", CallbackData: fmt.Sprintf("m04_10_%d", priceagent.ID)},
			},
			{
				{Text: "❌ Löschen", CallbackData: fmt.Sprintf("m04_99_%d", priceagent.ID)},
				{Text: "↩️ Zurück", CallbackData: backCallbackData},
			},
		},
	}
	_, err := cb.Message.EditText(b, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("showPriceagent: failed to edit message text: %w", err)
	}
	return nil
}

// changePriceagentSettingsHandler handles the callbacks for the buttons to change the notification
// settings of a price agent.
func changePriceagentSettingsHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	priceagentID, parseErr := parseIDFromCallbackData(cb.Data, "m04_00_")
	if parseErr != nil {
		return fmt.Errorf("changePriceagentSettingsHandler: failed to parse priceagentID from callback data: %w", parseErr)
	}

	priceagent, dbErr := database.GetPriceagentForUserByID(ctx.EffectiveUser.Id, priceagentID)
	if dbErr != nil {
		return fmt.Errorf("setNotifPriceagentHandler: failed to get priceagent from database: %w", dbErr)
	}

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("changePriceagentSettingsHandler: failed to answer callback query: %w", err)
	}

	linkName := createLink(priceagent.Entity.URL, priceagent.Entity.Name)
	editedText := fmt.Sprintf("Wann möchtest du für %s alarmiert werden?\nAktuelle Einstellung: %s\n\nAktueller Preis: %s", linkName, bold(priceagent.NotificationSettings.String()), bold(createPrice(priceagent.Entity.Price)))
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
	_, err := cb.Message.EditText(b, editedText, &gotgbot.EditMessageTextOpts{ReplyMarkup: markup, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("showPriceagent: failed to edit message text: %w", err)
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
	priceagentID, parseErr := parseIDFromCallbackData(cb.Data, "m04_99_")
	if parseErr != nil {
		return fmt.Errorf("deletePriceagentHandler: failed to parse priceagentID from callback data: %w", parseErr)
	}

	priceagent, dbErr := database.GetPriceagentForUserByID(ctx.EffectiveUser.Id, priceagentID)
	if dbErr != nil {
		ctx.EffectiveMessage.Reply(b, "Der Preisagent existiert nicht mehr, vielleicht wurde er schon gelöscht?", &gotgbot.SendMessageOpts{})
		return fmt.Errorf("deletePriceagentHandler: failed to get priceagent from database: %w", dbErr)
	}

	deleteErr := database.DeletePriceAgentForUser(priceagent)
	if deleteErr != nil {
		ctx.EffectiveMessage.Reply(b, "Der Preisagent konnte nicht gelöscht werden!", &gotgbot.SendMessageOpts{})
		return fmt.Errorf("deletePriceagentHandler: failed to delete priceagent from database: %w", deleteErr)
	}

	editText := fmt.Sprintf("Preisagent für %s wurde gelöscht!", bold(priceagent.Entity.Name))
	undoKeyboardMarkup := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "↩️ Rückgängig", CallbackData: fmt.Sprintf("m04_98_%d", priceagentID)},
			},
		},
	}
	_, err := cb.Message.EditText(b, editText, &gotgbot.EditMessageTextOpts{ReplyMarkup: undoKeyboardMarkup, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("deletePriceagent: failed to edit message text: %w", err)
	}
	return nil
}

func newUserHandler(_ *gotgbot.Bot, ctx *ext.Context) error {
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

func cbqNotImplementedHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Not implemented yet"}); err != nil {
		return fmt.Errorf("priceHistoryHandler: failed to answer callback query: %w", err)
	}
	return nil
}

// setCommands sets all the available commands for the bot on Telegram
func setCommands() {
	_, setCommandErr := bot.SetMyCommands([]gotgbot.BotCommand{
		{Command: "start", Description: "Startmenü des Bots"},
		{Command: "help", Description: "Zeigt die Hilfe an"},
		{Command: "show", Description: "Zeigt deine Preisagenten an"},
		{Command: "new", Description: "Fügt neuen Preisagenten hinzu"},
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
	dispatcher.AddHandler(handlers.NewCommand("version", versionHandler))
	dispatcher.AddHandler(handlers.NewCommand("help", helpHandler))

	// Callback Queries (inline keyboards)
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_98_"), cbqNotImplementedHandler)) // undo delete

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_99_"), deletePriceagentHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_10_"), priceHistoryHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_02_"), setNotificationBelowHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_01_"), setNotificationAlwaysHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_00_"), changePriceagentSettingsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m03_00_"), showPriceagentDetail))
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
		Client:      http.Client{},
		GetTimeout:  gotgbot.DefaultGetTimeout,
		PostTimeout: gotgbot.DefaultPostTimeout,
	})
	if createBotErr != nil {
		log.Println(botConfig)
		log.Fatalln("Something wrong:", createBotErr)
	}

	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: nil,
		DispatcherOpts: ext.DispatcherOpts{
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				fmt.Println("an error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			Panic:       nil,
			ErrorLog:    nil,
			MaxRoutines: 0,
		},
	})

	addMessageHandlers(updater.Dispatcher)
	setCommands()

	if botConfig.Webhook.Enabled {
		parsedURL, parseErr := url.Parse(botConfig.Webhook.URL)
		if parseErr != nil {
			log.Fatalln("Can't parse webhook url:", parseErr)
		}
		log.Println("Starting webhook...")
		// TODO add support for custom certificates
		err := updater.StartWebhook(bot, ext.WebhookOpts{
			Listen:  botConfig.Webhook.ListenIP,
			Port:    botConfig.Webhook.ListenPort,
			URLPath: parsedURL.Path,
		})
		bot.SetWebhook(botConfig.Webhook.URL, &gotgbot.SetWebhookOpts{})
		if err != nil {
			panic("failed to start webhook: " + err.Error())
		}
	} else {
		log.Println("Start polling...")
		_, _ = bot.DeleteWebhook(nil)
		err := updater.StartPolling(bot, &ext.PollingOpts{DropPendingUpdates: false})
		if err != nil {
			panic("failed to start polling: " + err.Error())
		}
	}
	fmt.Printf("Bot has been started as @%s...\n", bot.User.Username)

	updater.Idle()
}
