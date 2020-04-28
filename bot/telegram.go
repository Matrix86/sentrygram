package bot

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/Matrix86/sentrygram/pluginmanager"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Matrix86/sentrygram/utils"

	"github.com/evilsocket/islazy/async"
	"github.com/evilsocket/islazy/log"
	"gopkg.in/telegram-bot-api.v4"
)

type Telegram struct {
	bot          *tgbotapi.BotAPI
	api          string
	debug        bool
	enabledUsers sync.Map
	cacheId      sync.Map
	profilePath  string

	queue *async.WorkQueue
	msgcb msgCallback
	quit  chan int
}

type CallbackArgument struct {
	T   *Telegram
	Msg *tgbotapi.Message
}

type msgCallback func(msg *tgbotapi.Message, bot *Telegram) error

func NewTelegram(api string, users []string, timeout int, cb msgCallback, debug bool, profilePath string) (Bot, error) {
	var err error
	t := &Telegram{}

	if t.bot, err = tgbotapi.NewBotAPI(api); err != nil {
		return nil, fmt.Errorf("NewTelegram: %s", err)
	}

	t.api = api
	t.debug = debug
	t.quit = make(chan int)
	t.msgcb = cb
	t.profilePath = profilePath
	t.loadUsernames()

	for _, n := range users {
		t.enabledUsers.Store(n, true)
	}

	// Bot's methods usable on JS
	defines := map[string]interface{}{
		"sendMessage": func(to string, message string) interface{} {
			if err := t.SendMessage(to, message); err != nil {
				log.Error("sendMessage: %s", err)
			}
			return err
		},
		"sendImage": func(to string, path string) interface{} {
			if err := t.SendImage(to, path); err != nil {
				log.Error("sendImage: %s", err)
			}
			return err
		},
		"sendFile": func(to string, path string) interface{} {
			if err := t.SendFile(to, path); err != nil {
				log.Error("sendFile: %s", err)
			}
			return err
		},
		"sendAudio": func(to string, path string) interface{} {
			if err := t.SendAudio(to, path); err != nil {
				log.Error("SendAudio: %s", err)
			}
			return err
		},
		"sendVideo": func(to string, path string) interface{} {
			if err := t.SendVideo(to, path); err != nil {
				log.Error("SendVideo: %s", err)
			}
			return err
		},
		"addBotAdmin": func(s string) interface{} {
			t.enabledUsers.Store(s, true)
			return nil
		},
		"getBotAdmins": func() interface{} {
			list := make([]string, 0)
			t.enabledUsers.Range(func(key interface{}, value interface{}) bool {
				list = append(list, key.(string))
				return true
			})
			return list
		},
		"kickUser": func(username string, chat string) interface{} {
			chatID, err := t.getUserId(chat)
			if err != nil {
				return false
			}
			userID, err := t.getUserId(username)
			if err != nil {
				return false
			}
			cfg := tgbotapi.KickChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: chatID,
					UserID: int(userID),
				},
			}
			_, err = t.bot.KickChatMember(cfg)
			if err != nil {
				log.Error("kickUser: %s", err)
				return false
			}
			return true
		},
		"leaveGroup": func(chat string) interface{} {
			chatID, err := t.getUserId(chat)
			if err != nil {
				return false
			}
			cfg := tgbotapi.ChatConfig{
				ChatID: chatID,
			}
			_, err = t.bot.LeaveChat(cfg)
			if err != nil {
				log.Error("leaveGroup: %s", err)
				return false
			}
			return true
		},
		"leaveGroupById": func(sChatID string) interface{} {
			chatID, err := strconv.ParseInt(sChatID, 10, 64)
			if err != nil {
				log.Error("leaveGroupById: %s", err)
				return false
			}
			cfg := tgbotapi.ChatConfig{
				ChatID: chatID,
			}
			_, err = t.bot.LeaveChat(cfg)
			if err != nil {
				log.Error("leaveGroupById: %s", err)
				return false
			}
			return true
		},
		"getChatMembersCount": func(chat string) interface{} {
			chatID, err := t.getUserId(chat)
			if err != nil {
				return false
			}
			cfg := tgbotapi.ChatConfig{
				ChatID: chatID,
			}
			num, err := t.bot.GetChatMembersCount(cfg)
			if err != nil {
				log.Error("getChatMembersCount: %s", err)
				return -1
			}
			return num
		},
		"getCachedIds": func() interface{} {
			tmpMap := make(map[string]int64)
			t.cacheId.Range(func(key interface{}, value interface{}) bool {
				var username string
				var ok bool
				var id int64
				if username, ok = key.(string); !ok {
					log.Error("error on username")
					return true
				}
				if id, ok = value.(int64); !ok {
					log.Error("error on id")
					return true
				}
				tmpMap[username] = id
				return true
			})
			return tmpMap
		},
	}

	pm := pluginmanager.GetInstance()
	pm.SetDefines(defines)

	return t, nil
}

func (t *Telegram) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Info("Logged on account %s", t.bot.Self.UserName)

	updates, err := t.bot.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("Run(): %s", err)
	}

	t.queue = async.NewQueue(0, func(arg async.Job) {
		if arg != nil {
			args := arg.(*CallbackArgument)

			if args.T.msgcb != nil {
				// Handle message
				if err := args.T.msgcb(args.Msg, t); err != nil {
					log.Error("callback error: %s", err)
				}
			}
		}
	})
	defer t.queue.WaitDone()

	utils.DoEvery(1*time.Minute, t.flushUsernames)

	for {
		select {
		case update := <-updates:
			if update.Message != nil {
				t.queue.Add(async.Job(&CallbackArgument{T: t, Msg: update.Message}))
			} else if update.ChannelPost != nil {
				log.Debug("Received %#v", update.ChannelPost)
				t.queue.Add(async.Job(&CallbackArgument{T: t, Msg: update.ChannelPost}))
			}

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
	if id, ok := t.cacheId.Load(username); ok {
		msg := tgbotapi.NewMessage(id.(int64), text)
		msg.ParseMode = "html"
		if _, err := t.bot.Send(msg); err != nil {
			log.Error("SendMessage: %s", err)
			return err
		}
	} else {
		return errors.New("SendMessage: Username not found")
	}

	return nil
}

func (t *Telegram) SendImage(username string, path string) error {
	if id, ok := t.cacheId.Load(username); ok {
		msg := tgbotapi.NewPhotoUpload(id.(int64), path)
		if _, err := t.bot.Send(msg); err != nil {
			log.Error("sendImage: %s", err)
			return err
		}
	} else {
		return errors.New("sendImage: Username not found")
	}

	return nil
}

func (t *Telegram) SendFile(username string, path string) error {
	if id, ok := t.cacheId.Load(username); ok {
		msg := tgbotapi.NewDocumentUpload(id.(int64), path)
		if _, err := t.bot.Send(msg); err != nil {
			log.Error("SendFile: %s", err)
			return err
		}
	} else {
		return errors.New("SendFile: Username not found")
	}

	return nil
}

func (t *Telegram) SendAudio(username string, path string) error {
	if id, ok := t.cacheId.Load(username); ok {
		msg := tgbotapi.NewAudioUpload(id.(int64), path)
		if _, err := t.bot.Send(msg); err != nil {
			log.Error("SendAudio: %s", err)
			return err
		}
	} else {
		return errors.New("SendAudio: Username not found")
	}

	return nil
}

func (t *Telegram) SendVideo(username string, path string) error {
	if id, ok := t.cacheId.Load(username); ok {
		msg := tgbotapi.NewVideoUpload(id.(int64), path)
		if _, err := t.bot.Send(msg); err != nil {
			log.Error("SendVideo: %s", err)
			return err
		}
	} else {
		return errors.New("SendVideo: Username not found")
	}

	return nil
}

func (t *Telegram) flushUsernames(time time.Time) {
	tmpMap := make(map[string]int64)
	if t.profilePath == "" {
		return
	}

	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	t.cacheId.Range(func(key interface{}, value interface{}) bool {
		var username string
		var ok bool
		var id int64
		if username, ok = key.(string); !ok {
			log.Error("error on username")
			return true
		}
		if id, ok = value.(int64); !ok {
			log.Error("error on id")
			return true
		}
		tmpMap[username] = id
		return true
	})

	if err := e.Encode(tmpMap); err != nil {
		log.Error("flushUsername Encode: %s", err)
	}

	if err := ioutil.WriteFile(path.Join(t.profilePath, "ids_cache.dat"), b.Bytes(), 0644); err != nil {
		log.Error("flushUsername WriteFile: %s", err)
	}
}

func (t *Telegram) loadUsernames() bool {
	if t.profilePath == "" {
		return false
	}

	dat, err := ioutil.ReadFile(path.Join(t.profilePath, "ids_cache.dat"))
	if err != nil {
		log.Error("loadUsernames: ReadFile: %s", err)
		return false
	}

	var tmpMap map[string]int64
	b := bytes.NewBuffer(dat)
	d := gob.NewDecoder(b)
	if err = d.Decode(&tmpMap); err != nil {
		log.Error("loadUsernames: Decode: %s", err)
		return false
	}

	// cannot directly unmarshal into a sync.Map obj
	for u, i := range tmpMap {
		t.cacheId.Store(u, i)
	}

	return true
}

func (t *Telegram) GetCommandArgs(cmd string) (string, []string, error) {
	if !strings.HasPrefix(cmd, "/") {
		return "", nil, fmt.Errorf("not a command")
	}

	// Remove '/' char
	cmd = cmd[1:]
	r := csv.NewReader(strings.NewReader(cmd))
	r.Comma = ' '
	args, err := r.Read()
	if err != nil {
		return "", nil, fmt.Errorf("GetCommandArgs: %s", err)
	}

	cmd = args[0]

	return strings.Title(cmd), args[1:], nil
}

func (t *Telegram) getUserId(username string) (int64, error) {
	if id, ok := t.cacheId.Load(username); ok {
		return id.(int64), nil
	}
	return 0, fmt.Errorf("username %s not found", username)
}

func (t *Telegram) CacheUsername(username string, id int64) bool {
	if _, ok := t.cacheId.Load(username); !ok {
		t.cacheId.Store(username, id)
		return true
	}
	return false
}
