package pluginmanager

import (
	"fmt"
	"sync"
	"unicode"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/plugin"
)

type PluginManager struct {
	path        string
	commands    map[string]*plugin.Plugin
	initialized bool
}

var (
	instance *PluginManager
	once     sync.Once
)

func GetInstance() *PluginManager {
	once.Do(func() {
		instance = &PluginManager{
			initialized: false,
			commands:    make(map[string]*plugin.Plugin),
		}
	})
	return instance
}

func (pm *PluginManager) LoadPlugins(path string) error {
	if fs.Exists(path) == false {
		return fmt.Errorf("LoadPlugins: directory '%s' not found", path)
	}

	pm.SetInitialDefines()

	err := fs.Glob(path, "*.js", func(filename string) error {
		log.Debug("Plugin '%s' founded", filename)
		p, err := plugin.Load(filename)
		if err != nil {
			return err
		}
		for _, m := range p.Methods() {
			if unicode.IsUpper(rune(m[0])) {
				log.Debug("Method found: %s", m)
				pm.commands[m] = p
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("LoadPlugins: fs.Glob: %s", err)
	}
	pm.path = path
	pm.initialized = true

	return nil
}

func (pm *PluginManager) IsInitialized() bool {
	return pm.initialized
}

func (pm *PluginManager) GetCommands() []string {
	commands := make([]string, 0)
	if pm.IsInitialized() {
		for key, _ := range pm.commands {
			commands = append(commands, key)
		}
	}
	return commands
}

func (pm *PluginManager) HasCommand(cmd string) bool {
	pm.GetCommands()
	if pm.IsInitialized() {
		if len(cmd) == 0 {
			return false
		}
		if _, ok := pm.commands[cmd]; ok {
			return true
		}
	}
	return false
}

func (pm *PluginManager) Exec(cmd string, req interface{}) (interface{}, error) {
	if !pm.IsInitialized() {
		return "", fmt.Errorf("Exec: not initialized")
	} else if pm.HasCommand(cmd) {
		if ret, err := pm.commands[cmd].Call(cmd, req); err != nil {
			return nil, fmt.Errorf("Exec: %s", err)
		} else {
			return ret, nil
		}
	} else {
		return "", fmt.Errorf("Exec: command '%s' not found", cmd)
	}
}
