package bot

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// versionHandler handles the /version command.
func versionHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	_, sendErr := b.SendMessage(ctx.EffectiveChat.Id, "Bot version: "+version, nil)
	return sendErr
}
