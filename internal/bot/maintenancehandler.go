package bot

import (
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// maintenanceHandler
func maintenanceHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	maintenanceText := fmt.Sprintf("Hallo lieber Nutzer und vielen Dank für dein Interesse an diesem Bot.\n\nVor einiger Zeit wurde die Geizhals Webseite umgestaltet, wodurch der Bot keine Daten mehr auslesen kann. Aktuell schaffe ich es zeitlich nicht, den Bot an die neue Seite anzupassen. Diese Anpassung kommt mit erheblichem Aufwand daher, weshalb ich den Bot erst einmal deaktiviert habe.\n\nWenn du programmieren kannst, schau dir gerne den %s an - vielleicht hast du gerade etwas Zeit zum unterstützen?\n\nVielen Dank für dein Verständnis!", createLink("https://github.com/d-Rickyy-b/GoGeizhalsBot", "Quellcode auf GitHub"))
	_, replyErr := ctx.EffectiveMessage.Reply(bot, maintenanceText, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
		LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	if replyErr != nil {
		log.Println(replyErr)
	}

	return nil
}
