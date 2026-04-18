/* Package config provides configuration injection with hot reloads.

The configuration hot reloading requires that a pointer is returned to the underlying configuration. This allows for altering the configuration however, that should be avoided as the hot reload will overwrite any local changes. Please treat the returned configuration variable as immutable.
*/

package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Initializer interface provides an Initialize method allowing for set up of an object once it's configuration has been loaded
type Initializer interface {
	Initialize()
}

type (
	configMapType  map[string][]byte
	configPtrsType map[string]any
	initMapType    map[string]Initializer
	config         struct {
		lastLoad     time.Time
		configs      configMapType
		configPtrs   configPtrsType
		initializers initMapType
		configFile   string
		initialLoad  bool
	}
)

const configFileString = "PROJECT_CONFIG_FILE"

//nolint:gochecknoglobals // cfg holds the configuration data, which should only be retrieved via getConfig()
var (
	cfg               *config
	lock              = &sync.Mutex{}
	once              sync.Once
	hotReloadStop     chan struct{}
	hotReloadDisabled bool
)

//nolint:gochecknoinits // cfg is a singleton of configs; this ensures that it is initialized properly
func init() {
	cfg = &config{}

	cfg.configs = make(configMapType)
	cfg.configPtrs = make(configPtrsType)
	cfg.initializers = make(initMapType)
	cfg.lastLoad = time.Now()
	cfg.initialLoad = true
}

// getConfigFile returns the full path and filename of the configuration file
func (c config) getConfigFile() string {
	if c.configFile == "" {
		// TODO: better way to get config file location
		if filename := os.Getenv(configFileString); filename == "" {
			panic(fmt.Errorf("env var %s is not set", configFileString))
		} else {
			c.configFile = filename
		}
	}
	return c.configFile
}

// getConfig returns the internal cfg object (loading it if needed)
func getConfig() *config {
	if cfg.initialLoad {
		if err := load(); err != nil {
			panic(err)
		}
		cfg.initialLoad = false
	}

	return cfg
}

func load() error {
	lock.Lock()
	defer func() {
		cfg.lastLoad = time.Now()
		lock.Unlock()
	}()

	rawData, err := os.ReadFile(cfg.getConfigFile())
	if err != nil {
		return fmt.Errorf("unable to open config file (%s): %s", cfg.getConfigFile(), err)
	}

	if !json.Valid(rawData) {
		return fmt.Errorf("invalid JSON found in file (%s)", cfg.getConfigFile())
	}

	data := map[string]any{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		return fmt.Errorf("unable to unmarshal config file (%s): %s", cfg.getConfigFile(), err)
	}

	for key, value := range data {
		valueJson, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("unable to marshal key (%s) data: %s", key, err)
		}

		if _, ok := cfg.configPtrs[key]; !ok {
			cfg.configs[key] = valueJson
		} else {
			err := json.Unmarshal(valueJson, cfg.configPtrs[key])
			if err != nil {
				return fmt.Errorf("unable to unmarshal pointer for %s: %s", key, err)
			}
		}

		if _, ok := cfg.initializers[key]; ok {
			cfg.initializers[key].Initialize()
		}
	}

	return nil
}

const checkInterval = 15 // seconds

// DisableHotReload prevents the config file from being watched for changes.
// If called before LoadConfig, the watcher goroutine is never started.
// If called after, the existing watcher is stopped.
func DisableHotReload() {
	hotReloadDisabled = true
	if hotReloadStop != nil {
		close(hotReloadStop)
	}
}

func startHotReload() {
	if hotReloadDisabled {
		return
	}
	once.Do(func() {
		hotReloadStop = make(chan struct{})
		ticker := time.NewTicker(checkInterval * time.Second)
		go func() {
			for {
				select {
				case <-hotReloadStop:
					ticker.Stop()
					return
				case <-ticker.C:
					fileInfo, err := os.Stat(cfg.getConfigFile())
					if err != nil {
						log.Printf("config reload: unable to stat file (%s): %s", cfg.getConfigFile(), err)
						continue
					}
					if fileInfo.ModTime().Sub(cfg.lastLoad) > 0 {
						if err := load(); err != nil {
							log.Printf("config reload: %s", err)
						}
					}
				}
			}
		}()
	})
}

// LoadConfig takes a string (which matches one of the top level JSON keys in the config) and a
// reference to a struct that will be populated with the config data.
//
// This function also sets up a check of the config file for any modifications. If changes are detected the config will be
// reloaded. Any errors encountered during a reload are logged and the previous configuration is retained.
// Hot reloading can be disabled by calling DisableHotReload.
func LoadConfig(name string, configStruct any) {
	cfg = getConfig()
	startHotReload()

	cfgPtr, ok := cfg.configPtrs[name]
	if ok {
		configStruct = cfgPtr
	} else {
		configData, ok := cfg.configs[name]
		if !ok {
			panic(fmt.Errorf("key (%s) not found in config file (%s)", name, cfg.getConfigFile()))
		}
		err := json.Unmarshal(configData, &configStruct)
		if err != nil {
			panic(fmt.Errorf("unable to unmarshal (%s) to struct: %s", configData, err))
		}
		cfg.configPtrs[name] = &configStruct
	}
}

// RegisterInitializer 'registers' the struct as being able to be initialized and runs that routine.
//
//	TODO: this should be replaced as this should be done programmatically
func RegisterInitializer(name string, initFunc Initializer) {
	cfg.initializers[name] = initFunc
	initFunc.Initialize()
}
