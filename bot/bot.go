package bot

type Bot interface {
	SendMessage(username string, text string) error
	SendImage(username string, path string) error
	Run() error
	Stop()
}

type BotMessage struct {
	From         string
	Content      interface{}
	ChatName     string
	IsGroup      bool
	IsPrivate    bool
	IsSuperGroup bool
	IsChannel    bool
	IsCommand    bool
}
