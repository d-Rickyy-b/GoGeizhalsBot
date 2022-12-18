package bot

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// helpHandler handles all the messages containing the /help command.
func helpHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	helpMessage := "Du brauchst Hilfe? Probiere folgende Befehle:\n" +
		"\n" +
		"/start - Startmenü\n" +
		"/help - Zeigt diese Hilfe\n" +
		"/stop - Löscht alle deine Daten und beendet den Bot\n" +
		"/version - Zeigt die aktuelle Version des Bots"
	_, err := ctx.Message.Reply(b, helpMessage, nil)

	return err
}
