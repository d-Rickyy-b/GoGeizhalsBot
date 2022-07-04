package bot

import (
	"GoGeizhalsBot/internal/database"
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// stopHandler handles the /stop command. It sends the user a message with an inline keyboard
// to confirm that the user actually wants to stop the bot and delete all data.
func stopHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	// Ask user if they want to stop the bot and delete all their data
	// markup contains a replykeyboardmarkup with two buttons "Stop bot" and "Cancel"
	areYouSureText := fmt.Sprintf("%s\n\n%s", bold("Möchtest du den Bot wirklich stoppen?"), "Dadurch werden alle deine Daten (& Preisagenten) gelöscht und der Bot wird dich nicht mehr benachrichtigen!")
	_, replyErr := ctx.EffectiveMessage.Reply(b, areYouSureText, &gotgbot.SendMessageOpts{
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

func stopHandlerConfirm(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	database.DeleteUserWithCache(cb.From.Id)
	_, editErr := cb.Message.EditText(b, "Deine Daten wurden gelöscht!", &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
	})
	if editErr != nil {
		log.Println(editErr)
	}
	return nil
}

func stopHandlerCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	_, editErr := cb.Message.EditText(b, "Der Vorgang wurde abgebrochen! Deine Daten wurden nicht gelöscht.", &gotgbot.EditMessageTextOpts{
		ParseMode: "HTML",
	})
	if editErr != nil {
		log.Println(editErr)
	}

	return nil
}
