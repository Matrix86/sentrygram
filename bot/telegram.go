package bot

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/evilsocket/islazy/async"
	"github.com/evilsocket/islazy/log"
	"gopkg.in/telegram-bot-api.v4"
)

type Telegram struct {
	bot          *tgbotapi.BotAPI
	api          string
	debug        bool
	enabledUsers []string
	chatId       map[string]int64
	profilePath  string

	queue *async.WorkQueue
	msgcb msgCallback
	quit  chan int
}

type CallbackArgument struct {
	T   *Telegram
	Msg *tgbotapi.Message
}

const (
	Text int = iota
	Image
)

type TelegramMessage struct {
	ID      int64
	Content interface{}
	Type    int
}

type msgCallback func(msg *tgbotapi.Message, bot *Telegram) (TelegramMessage, error)

func NewTelegram(api string, users []string, timeout int, cb msgCallback, debug bool, profilePath string) (*Telegram, error) {
	var err error
	t := &Telegram{}

	if t.bot, err = tgbotapi.NewBotAPI(api); err != nil {
		return nil, fmt.Errorf("NewTelegram: %s", err)
	}

	t.api = api
	t.debug = debug
	t.quit = make(chan int)
	t.enabledUsers = users
	t.msgcb = cb
	t.profilePath = profilePath
	if !t.loadUsernames() {
		t.chatId = make(map[string]int64, 0)
	}

	return t, nil
}

func (t *Telegram) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := t.bot.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("Run(): %s", err)
	}

	t.queue = async.NewQueue(0, func(arg async.Job) {
		if arg != nil {
			args := arg.(*CallbackArgument)

			if args.T.msgcb != nil {
				// Handle message
				if newMsg, err := args.T.msgcb(args.Msg, t); err == nil {
					switch newMsg.Type {
					case Text:
						if text, ok := newMsg.Content.(string); ok {
							msg := tgbotapi.NewMessage(newMsg.ID, text)
							msg.ParseMode = "html"
							if _, err := args.T.bot.Send(msg); err != nil {
								log.Error("telegram callback: %s", err)
							}
						} else {
							log.Debug("callback didn't return a string")
						}

					case Image:
						// file is a string path to the file, FileReader, or FileBytes.
						msg := tgbotapi.NewPhotoUpload(newMsg.ID, newMsg.Content)
						args.T.bot.Send(msg)
					}
				} else {
					log.Debug("callback error: %s", err)
				}
			}
		}
	})
	defer t.queue.WaitDone()

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			t.queue.Add(async.Job(&CallbackArgument{T: t, Msg: update.Message}))

		case <-t.quit:
			return nil
		}
	}

	return nil
}

func (t *Telegram) Stop() {
	t.queue.Stop()
	t.quit <- 0
}

func (t *Telegram) SendMessage(username string, text string) error {
	if id, ok := t.chatId[username]; ok {
		msg := tgbotapi.NewMessage(id, text)
		if _, err := t.bot.Send(msg); err != nil {
			log.Error("telegram callback: %s", err)
			return err
		}
	} else {
		return errors.New("SendMessage: Username not found")
	}

	return nil
}

func (t *Telegram) flushUsernames() {
	if t.profilePath == "" {
		return
	}

	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	if err := e.Encode(t.chatId); err != nil {
		log.Error("flushUsername Encode: %s", err)
	}

	if err := ioutil.WriteFile(t.profilePath+"/chat_ids.json", b.Bytes(), 0644); err != nil {
		log.Error("flushUsername WriteFile: %s", err)
	}
}

func (t *Telegram) loadUsernames() bool {
	if t.profilePath == "" {
		return false
	}

	dat, err := ioutil.ReadFile(t.profilePath + "/chat_ids.json")
	if err != nil {
		log.Error("loadUsernames: ReadFile: %s", err)
		return false
	}

	b := bytes.NewBuffer(dat)
	d := gob.NewDecoder(b)
	if err = d.Decode(&t.chatId); err != nil {
		log.Error("loadUsernames: Decode: %s", err)
		return false
	}

	// Check if all the previous enabled users are still enabled
	for n, _ := range t.chatId {
		exists := false
		for _, n2 := range t.enabledUsers {
			if n2 == n {
				exists = true
				break
			}
		}
		if !exists {
			delete(t.chatId, n)
			t.flushUsernames()
		}
	}

	return true
}

func (t *Telegram) GetCommandArgs(cmd string) (string, string) {
	if !strings.HasPrefix(cmd, "/") {
		return "", ""
	}

	// Remove '/' char
	cmd = cmd[1:]
	args := ""
	if strings.Contains(cmd, " ") {
		i := strings.Index(cmd, " ")
		args = cmd[i+1:]
		cmd = cmd[0:i]
	}

	return strings.Title(cmd), args
}
