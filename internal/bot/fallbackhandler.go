package bot

import (
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func fallbackCallbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	log.Printf("fallbackCallbackHandler - handled data: %s\n", cb.Data)

	if _, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{}); err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	return nil
}

// fallbackCommandHandler handles messates with unknown commands. It does not reply to the user.
func fallbackCommandHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println(fmt.Sprintf("User sent unimplemented command: %s", ctx.Message.Text))
	return nil
}
