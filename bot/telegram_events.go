package bot

import (
	"github.com/Matrix86/sentrygram/core/pluginmanager"

	"github.com/evilsocket/islazy/log"
	"gopkg.in/telegram-bot-api.v4"
)

func OnMessage(msg *tgbotapi.Message, bot *Telegram) (TelegramMessage, error) {
	// Check if the user has permissions
	userIsEnabled := false
	for _, u := range bot.enabledUsers {
		if msg.Chat.UserName == u {
			userIsEnabled = true
			if _, ok := bot.chatId[u]; !ok {
				bot.chatId[u] = msg.Chat.ID
				bot.flushUsernames()
			}
			break
		}
	}

	// Handle text messages that contain commands
	if !userIsEnabled {
		log.Info("Received Message not authorized [%s : '%s']", msg.From.UserName, msg.Text)
		return TelegramMessage{ID: msg.Chat.ID, Content: "User not enabled"}, nil
	} else if msg.IsCommand() {
		log.Debug("Received Command [%s : '%s']", msg.Chat.UserName, msg.Text)
		cmd, _ := bot.GetCommandArgs(msg.Text)
		pm := pluginmanager.GetInstance()
		if pm.IsInitialized() {
			// Help command returns list of available commands
			if cmd == "Help" {
				commands := pm.GetCommands()
				txt := ""
				for _, c := range commands {
					txt += "/" + c + "\n"
				}
				retMsg := TelegramMessage{ID: msg.Chat.ID, Content: txt, Type: Text}
				return retMsg, nil
			}

			recMsg := TelegramMessage{ID: msg.Chat.ID, Content: msg.Text, Type: Text}
			retMsg := TelegramMessage{ID: msg.Chat.ID, Content: msg.Text, Type: Text}
			_, err := pm.Exec(cmd, &recMsg, &retMsg)
			if err != nil {
				log.Error("Received Command not recognized : '%s' : %s", msg.Text, err)
				return TelegramMessage{ID: msg.Chat.ID, Content: "Sorry I don't recognize this command"}, nil
			}
			return retMsg, nil
		} else {
			log.Error("bot.OnMessage: plugin manager is not initialized")
		}
	} else {
		log.Info("Received Message [%s : '%s']", msg.From.UserName, msg.Text)
	}

	return TelegramMessage{ID: msg.Chat.ID, Content: "Sorry I don't recognize this command"}, nil
}
