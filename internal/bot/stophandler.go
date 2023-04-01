package bot

import (
	"fmt"
	"log"

	"github.com/d-Rickyy-b/gogeizhalsbot/internal/database"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// stopHandler handles the /stop command. It sends the user a message with an inline keyboard
// to confirm that the user actually wants to stop the bot and delete all data.
func stopHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Ask user if they want to stop the bot and delete all their data
	// markup contains a replykeyboardmarkup with two buttons "Stop bot" and "Cancel"
	areYouSureText := fmt.Sprintf("%s\n\n%s", bold("Möchtest du den Bot wirklich stoppen?"), "Dadurch werden alle deine Daten (& Preisagenten) gelöscht und der Bot wird dich nicht mehr benachrichtigen!")
	_, replyErr := ctx.EffectiveMessage.Reply(bot, areYouSureText, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
				{Text: "⚠️ Daten löschen ⚠️", CallbackData: "m06_01"},
				{Text: "↩️ Abbrechen", CallbackData: "m06_02"},
			}},
		},
		ParseMode: "HTML",
	})

	if replyErr != nil {
		log.Println(replyErr)
	}

	return nil
}

func stopHandlerConfirm(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery
	database.DeleteUserWithCache(cbq.From.Id)
	_, _, editErr := cbq.Message.EditText(bot, "Deine Daten wurden gelöscht!", &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
	})

	if editErr != nil {
		log.Println(editErr)
	}

	return nil
}

func stopHandlerCancel(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery
	_, _, editErr := cbq.Message.EditText(bot, "Der Vorgang wurde abgebrochen! Deine Daten wurden nicht gelöscht.", &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
	})

	if editErr != nil {
		log.Println(editErr)
	}

	return nil
}
