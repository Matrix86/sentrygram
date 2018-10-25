package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/Matrix86/sentrygram/bot"
	"github.com/Matrix86/sentrygram/core"
	"github.com/Matrix86/sentrygram/core/pluginmanager"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"github.com/mitchellh/go-homedir"
)

var (
	profilePath string
	profileFile string
)

func main() {
	flag.StringVar(&profileFile, "p", "", "Path of the configuration.")

	flag.Parse()

	log.Output = ""
	log.Level = log.DEBUG
	log.OnFatal = log.ExitOnFatal
	log.Format = "[{datetime}] {level:color}{level:name}{reset} {message}"

	if len(profileFile) == 0 {
		homeDir, err := homedir.Dir()
		if err != nil {
			log.Fatal("%s", err)
		}

		profilePath = homeDir + "/." + core.Name
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			if err = os.Mkdir(profilePath, 0755); err != nil {
				log.Fatal("%s", err)
			}
		}
		profileFile = profilePath + "/main.conf"

		if !fs.Exists(profileFile) {
			log.Fatal("Configuration file not exists, please create it on '" + profilePath + " directory")
		}
	} else {
		profilePath = "."
	}

	config, err := core.LoadConfiguration(profileFile)
	if err != nil {
		log.Fatal("main.conf not found: %s", err)
		os.Exit(1)
	}

	log.Level = config.GetLogLevel()
	log.Output = config.Logs
	if err := log.Open(); err != nil {
		panic(err)
	}
	defer log.Close()

	manager := pluginmanager.GetInstance()
	err = manager.LoadPlugins(config.PluginsPath)
	if err != nil {
		log.Fatal("Error: %s", err)
	}

	bot, err := bot.NewTelegram(config.TgmAPI, config.Users, 60, bot.OnMessage, false, profilePath)
	if err != nil {
		log.Fatal("Error : %s\n", err)
		return
	}

	if config.RpcEnabled {
		go core.NewRpcHandler(config.RpcPort, bot)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Info("Captured %v, exiting..", sig)
			bot.Stop()
		}
	}()

	bot.Run()
}
