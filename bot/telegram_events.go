package bot

import (
	"fmt"

	"github.com/Matrix86/sentrygram/pluginmanager"

	"github.com/evilsocket/islazy/log"
	"gopkg.in/telegram-bot-api.v4"
)

var privateCommands = map[string]bool{
	"OnUserAddOnGroup": true,
	"OnGroupMessage": true,
	"OnChanMessage": true,
	"OnPrivateChatMessage": true,
	"OnSuperGroupMessage": true,
}

func OnMessage(msg *tgbotapi.Message, bot *Telegram) error {
	parseCommand := true

	// New members in the group
	if msg.NewChatMembers != nil {
		if _, err := OnUserAddOnGroup(msg, bot); err != nil {
			log.Error("OnUserAddOnGroup: %s", err)
		}
	}

	// store user ID if not present
	if msg.From != nil && msg.From.UserName != "" {
		bot.CacheUsername(msg.From.UserName, int64(msg.From.ID))
	}

	if msg.Chat.IsGroup() {
		bot.CacheUsername(msg.Chat.Title, msg.Chat.ID)
		ret, err := OnGroupMessage(msg, bot)
		if err != nil {
			log.Error("OnGroupMessage: %s", err)
		}
		parseCommand = parseCommand && ret
	} else if msg.Chat.IsChannel() {
		ret, err := OnChanMessage(msg, bot)
		if err != nil {
			log.Error("OnChanMessage: %s", err)
		}
		parseCommand = parseCommand && ret
	} else if msg.Chat.IsPrivate() {
		ret, err := OnPrivateChatMessage(msg, bot)
		if err != nil {
			log.Error("OnPrivateChatMessage: %s", err)
		}
		parseCommand = parseCommand && ret
	} else if msg.Chat.IsSuperGroup() {
		ret, err := OnSuperGroupMessage(msg, bot)
		if err != nil {
			log.Error("OnSuperGroupMessage: %s", err)
		}
		parseCommand = parseCommand && ret
	}

	if msg.Chat.IsChannel() == false {
		// Check if the user has permissions
		userIsEnabled := false
		if _, ok := bot.enabledUsers.Load(msg.From.UserName); ok {
			userIsEnabled = true
		}
		// Handle text messages that contain commands
		if parseCommand && userIsEnabled && msg.IsCommand() {
			return OnCommand(msg, bot)
		} else {
			log.Info("Received Message [%s : '%s']", msg.From.UserName, msg.Text)
		}
	}
	return nil
}

func OnGroupMessage(msg *tgbotapi.Message, bot *Telegram) (bool, error) {
	pm := pluginmanager.GetInstance()
	recMsg := BotMessage{From: msg.From.UserName, ChatName: msg.Chat.Title, Content: msg.Text, IsGroup: true, IsCommand: msg.IsCommand()}
	if pm.HasCommand("OnGroupMessage") {
		ret, err := pm.Exec("OnGroupMessage", &recMsg)
		if err != nil {
			return true, err
		}
		if r, ok := ret.(bool); ok {
			return r, nil
		}
	}
	return true, nil
}

func OnPrivateChatMessage(msg *tgbotapi.Message, bot *Telegram) (bool, error) {
	pm := pluginmanager.GetInstance()
	recMsg := BotMessage{From: msg.From.UserName, Content: msg.Text, IsPrivate: true, IsCommand: msg.IsCommand()}
	if pm.HasCommand("OnPrivateChatMessage") {
		ret, err := pm.Exec("OnPrivateChatMessage", &recMsg)
		if err != nil {
			return true, err
		}
		if r, ok := ret.(bool); ok {
			return r, nil
		}
	}
	return true, nil
}

func OnChanMessage(msg *tgbotapi.Message, bot *Telegram) (bool, error) {
	pm := pluginmanager.GetInstance()
	log.Debug("%#v", msg.Chat)
	recMsg := BotMessage{ChatName: msg.Chat.Title, Content: msg.Text, IsChannel: true, IsCommand: msg.IsCommand()}
	if pm.HasCommand("OnChanMessage") {
		ret, err := pm.Exec("OnChanMessage", &recMsg)
		if err != nil {
			return true, err
		}
		if r, ok := ret.(bool); ok {
			return r, nil
		}
	}
	return true, nil
}

func OnSuperGroupMessage(msg *tgbotapi.Message, bot *Telegram) (bool, error) {
	pm := pluginmanager.GetInstance()
	recMsg := BotMessage{From: msg.From.UserName, ChatName: msg.Chat.Title, Content: msg.Text, IsPrivate: true, IsCommand: msg.IsCommand()}
	if pm.HasCommand("OnSuperGroupMessage") {
		ret, err := pm.Exec("OnSuperGroupMessage", &recMsg)
		if err != nil {
			return true, err
		}
		if r, ok := ret.(bool); ok {
			return r, nil
		}
	}
	return true, nil
}

func OnCommand(msg *tgbotapi.Message, bot *Telegram) (error){
	log.Debug("Received Command [%s : '%s']", msg.From.UserName, msg.Text)
	cmd, _ := bot.GetCommandArgs(msg.Text)
	if _, ok := privateCommands[cmd]; ok {
		// This type of command is private!
		log.Error("Received private Command: '%s'", msg.Text)
		return bot.SendMessage(msg.From.UserName, fmt.Sprintf("Command '%s' not recognized", cmd))
	}

	pm := pluginmanager.GetInstance()
	if pm.IsInitialized() {
		// Help command returns list of available commands
		if cmd == "Help" {
			commands := pm.GetCommands()
			txt := ""
			for _, c := range commands {
				if _, ok := privateCommands[c]; !ok {
					txt += "/" + c + "\n"
				}
			}
			return bot.SendMessage(msg.From.UserName, txt)
		}

		recMsg := BotMessage{From: msg.From.UserName, Content: msg.Text, ChatName: msg.Chat.Title, IsPrivate: msg.Chat.IsPrivate(), IsGroup: msg.Chat.IsGroup(), IsSuperGroup: msg.Chat.IsSuperGroup(), IsChannel: msg.Chat.IsChannel()}
		_, err := pm.Exec(cmd, &recMsg)
		if err != nil {
			log.Error("Received Command not recognized : '%s' : %s", msg.Text, err)
			return bot.SendMessage(msg.From.UserName, fmt.Sprintf("Command '%s' not recognized", cmd))
		}
		return nil
	} else {
		log.Error("bot.OnCommand: plugin manager is not initialized")
		return bot.SendMessage(msg.From.UserName, "plugin manager is not initialized")
	}
}

func OnUserAddOnGroup(msg *tgbotapi.Message, bot *Telegram) (bool, error) {
	for _, u := range *msg.NewChatMembers {
		log.Debug("Added user '%s' to '%s'", u.UserName, msg.Chat.Title)
		bot.CacheUsername(u.UserName, int64(u.ID))
		pm := pluginmanager.GetInstance()
		recMsg := BotMessage{From: u.UserName, ChatName:  msg.Chat.Title, IsGroup: true}
		_, err := pm.Exec("OnUserAddOnGroup", &recMsg)
		if err != nil {
			return true, err
		}
	}
	return true, nil
}