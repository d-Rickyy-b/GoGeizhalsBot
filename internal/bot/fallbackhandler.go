package bot

import (
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// fallbackCallbackHandler logs all the callback queries that are not handled by the other handlers.
// That should barely ever happen.
func fallbackCallbackHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	cbq := ctx.Update.CallbackQuery
	log.Printf("fallbackCallbackHandler - handled data: %s\n", cbq.Data)

	if _, err := cbq.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	return nil
}

// fallbackCommandHandler handles messates with unknown commands. It does not reply to the user.
func fallbackCommandHandler(_ *gotgbot.Bot, ctx *ext.Context) error {
	log.Printf("User sent unimplemented command: %s\n", ctx.Message.Text)
	return nil
}
