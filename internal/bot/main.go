package bot

import (
	"GoGeizhalsBot/internal/bot/models"
	"GoGeizhalsBot/internal/bot/userstate"
	"GoGeizhalsBot/internal/config"
	"GoGeizhalsBot/internal/geizhals"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
)

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

	// TODO replace priceagents with actual subscribed priceagents
	// TODO add case for zero price agents

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
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	// TODO replace priceagents with actual subscribed priceagents
	// TODO add case for zero price agents

	markup := generateEntityKeyboard(priceagents, "m03", 2)
	_, err := cb.Message.EditText(b, "Das sind deine Preisagenten für deine Produkte:", &gotgbot.EditMessageTextOpts{ReplyMarkup: markup})
	if err != nil {
		return fmt.Errorf("showProduct: failed to edit message text: %w", err)
	}
	return nil
}

// newPriceagentHandler is a callback handler for the m01_00 callback.
func newPriceagentHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("newPriceagentHandler")
	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
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
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	// TODO get specific priceagent for user
	priceagent := getWishlistPriceagent()
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

// deletePriceagentHandler handles all the inline "delete" buttons for priceagents
func deletePriceagentHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	// TODO: Delete Priceagent from DB
	editText := fmt.Sprintf("Preisagent für %s wurde gelöscht!", bold("Sony WF-1000XM4 schwarz"))
	_, err := cb.Message.EditText(b, editText, &gotgbot.EditMessageTextOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{}, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("deletePriceagent: failed to edit message text: %w", err)
	}
	return nil
}

func fallbackCallbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	log.Printf("fallbackCallbackHandler - handled data: %s\n", cb.Data)

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	return nil
}

// textHandler handles all the text messages. It checks if the message contains a valid URL and if so, it creates a new pricehandler
func textHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Printf("User sent '%s'\n", ctx.EffectiveMessage.Text)

	// Check if text is a link to geizhals
	if !geizhals.IsValidURL(ctx.EffectiveMessage.Text) {
		log.Println("Message is not a valid geizhals URL!")
	}

	// Parse link and request price
	return nil
}

// Start is the main function to start the bot.
func Start() {
	var configFile = flag.String("config", "config.json", "Path to config file")
	flag.Parse()
	c, _ := config.ReadConfig(*configFile)

	b, err := gotgbot.NewBot(c.Token, &gotgbot.BotOpts{
		Client:      http.Client{},
		GetTimeout:  gotgbot.DefaultGetTimeout,
		PostTimeout: gotgbot.DefaultPostTimeout,
	})
	if err != nil {
		log.Println(c)
		log.Fatalln("Something wrong:", err)
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

	dispatcher := updater.Dispatcher
	dispatcher.AddHandler(handlers.NewCommand("start", startHandler))
	dispatcher.AddHandler(handlers.NewCommand("version", versionHandler))

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_10_"), cbqNotImplementedHandler)) // priceHistory
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_98_"), cbqNotImplementedHandler)) // undo delete

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_99_"), deletePriceagentHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_02_"), setNotificationBelowHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_01_"), setNotificationAlwaysHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m04_00_"), changePriceagentSettingsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("m03_00_"), showPriceagentDetail))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m02_00"), showWishlistPriceagents))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m02_01"), showProductPriceagents))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m01_01"), viewPriceagentsHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m01_00"), newPriceagentHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("m00_00"), mainMenuHandler))

	dispatcher.AddHandler(handlers.NewMessage(message.Text, textHandler))

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.All, fallbackCallbackHandler))

	err = updater.StartPolling(b, &ext.PollingOpts{DropPendingUpdates: false})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	fmt.Printf("%s has been started...\n", b.User.Username)

	updater.Idle()
}
