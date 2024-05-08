// Package config provides configuration injection with hot reloads.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Initializer interface provides an Initialize method allowing for set up of an object once it's configuration has been loaded
type Initializer interface {
	Initialize()
}

type configMapType map[string][]byte
type configPtrsType map[string]any
type initMapType map[string]Initializer
type config struct {
	configFile   string         // location of the config json file
	configs      configMapType  // map containing json representation of each top level key
	configPtrs   configPtrsType // map of pointers to existing config objects
	lastLoad     time.Time      // time of last config load
	reload       bool           // set to true if config file needs to be reloaded
	initializers initMapType    // initialization functions that some modules may need
}

const configFileString = "PROJECT_CONFIG_FILE"

var cfg *config

func init() {
	cfg = &config{}

	cfg.configs = make(configMapType)
	cfg.configPtrs = make(configPtrsType)
	cfg.initializers = make(initMapType)
	cfg.lastLoad = time.Now()
	cfg.reload = true
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

var lock = &sync.Mutex{}

// getConfig returns the internal cfg object (loading it if needed)
func getConfig() *config {
	if cfg.reload {
		load()
	}

	return cfg
}

// This method is usually called first and the application can not run without this information.  Since any errors
// encountered here are fatal, panic is used instead of any type of error return or logging.
func load() {
	lock.Lock()
	defer func() {
		cfg.lastLoad = time.Now()
		cfg.reload = false

		lock.Unlock()
	}()

	rawData, err := os.ReadFile(cfg.getConfigFile())
	if err != nil {
		panic(fmt.Errorf("unable to open config file (%s): %s", cfg.getConfigFile(), err))
	}

	if !json.Valid(rawData) {
		panic(fmt.Errorf("invalid JSON found in file (%s)", cfg.getConfigFile()))
	}

	data := map[string]interface{}{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal config file (%s): %s", cfg.getConfigFile(), err))
	}

	for key, value := range data {
		valueJson, err := json.Marshal(value)
		if err != nil {
			panic(fmt.Errorf("unable to marshal key (%s) data: %s", key, err))
		}

		if _, ok := cfg.configPtrs[key]; !ok {
			cfg.configs[key] = valueJson
		} else {
			// using Unmarshal to take care of all the reflection work
			err := json.Unmarshal(valueJson, cfg.configPtrs[key])
			if err != nil {
				panic(fmt.Errorf("unable to unmarshal pointer for %s: %s", key, err))
			}
		}

		if _, ok := cfg.initializers[key]; ok {
			cfg.initializers[key].Initialize()
		}
	}
}

// TODO: make this configurable
const checkInterval = 15 // seconds
var once sync.Once

// LoadConfig takes a string (which matches one of the top level JSON keys in the config) and a
// reference to a struct that will be populated with the config data.
//
// This function also sets up a check of the config file for any modifications. If changes are detected the config will be
// reloaded. Any errors encountered during the re-parsing of the config will terminate the program.
func LoadConfig(name string, configStruct interface{}) {
	cfg = getConfig()
	once.Do(func() {
		ticker := time.NewTicker(checkInterval * time.Second)
		go func() {
			for range ticker.C {
				fileInfo, err := os.Stat(cfg.getConfigFile())
				if err != nil {
					panic(fmt.Errorf("unable state file (%s): %s", cfg.getConfigFile(), err))
				}
				if fileInfo.ModTime().Sub(cfg.lastLoad) > 0 {
					load()
				}
			}
		}()
	})

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
