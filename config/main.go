/* Package config provides configuration injection with hot reloads.

The configuration hot reloading requires that a pointer is returned to the underlying configuration. This allows for altering the configuration however, that should be avoided as the hot reload will overwrite any local changes. Please treat the returned configuration variable as immutable.
*/

package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Initializer interface provides an Initialize method allowing for set up of an object once it's configuration has been loaded
type Initializer interface {
	Initialize()
}

type (
	configMapType  map[string][]byte
	configPtrsType map[string]any
	config         struct {
		lastLoad       time.Time
		configs        configMapType
		configPtrs     configPtrsType
		baseFile       string
		envFile        string
		envFileExisted bool
		initialLoad    bool
	}
)

const (
	configFileString          = "PROJECT_CONFIG_FILE"
	configDirString           = "PROJECT_CONFIG_DIR"
	configEnvVarString        = "PROJECT_CONFIG_ENV_VAR"
	configCheckIntervalString = "PROJECT_CONFIG_CHECK_INTERVAL"
	pathSeparator             = ":"
	defaultCheckInterval      = 15
)

//nolint:gochecknoglobals // cfg holds the configuration data, which should only be retrieved via getConfig()
var (
	cfg               *config
	lock              = &sync.Mutex{}
	hotReloadStart    sync.Once
	hotReloadStopOnce sync.Once
	hotReloadStop     chan struct{}
	hotReloadDisabled atomic.Bool
	checkInterval     = defaultCheckInterval //nolint:gochecknoglobals // set once in init()
)

//nolint:gochecknoinits // cfg is a singleton of configs; this ensures that it is initialized properly
func init() {
	cfg = &config{}

	cfg.configs = make(configMapType)
	cfg.configPtrs = make(configPtrsType)
	cfg.lastLoad = time.Now()
	cfg.initialLoad = true
	cfg.resolveConfigSources()

	if val := os.Getenv(configCheckIntervalString); val != "" {
		parsed, err := strconv.Atoi(val)
		if err != nil || parsed < 1 {
			panic(fmt.Errorf("env var %s must be a positive integer, got: %s", configCheckIntervalString, val))
		}
		checkInterval = parsed
	}
}

func (c *config) resolveConfigSources() {
	filename := os.Getenv(configFileString)
	if filename == "" {
		panic(fmt.Errorf("env var %s is not set", configFileString))
	}

	dirList := os.Getenv(configDirString)
	if dirList == "" {
		dirList = "."
	}

	for _, dir := range strings.Split(dirList, pathSeparator) {
		candidate := filepath.Join(dir, filename)
		if _, err := os.Stat(candidate); err == nil { //nolint:gosec // path from env var config
			c.baseFile = candidate
			break
		}
	}

	if c.baseFile == "" {
		panic(fmt.Errorf("%s (%s) not found in directories: %s", configFileString, filename, dirList))
	}

	if envVarName := os.Getenv(configEnvVarString); envVarName != "" {
		if env := os.Getenv(envVarName); env != "" {
			ext := filepath.Ext(filename)
			stem := strings.TrimSuffix(filename, ext)
			envFilename := fmt.Sprintf("%s_%s%s", stem, env, ext)
			c.envFile = filepath.Join(filepath.Dir(c.baseFile), envFilename)
		}
	}
}

func initializeIfSupported(v any) {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return
		}
		if init, ok := val.Interface().(Initializer); ok {
			init.Initialize()
			return
		}
		val = val.Elem()
	}
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

func readJSONFile(path string) (map[string]any, error) {
	rawData, err := os.ReadFile(path) //nolint:gosec // path resolved from env var config
	if err != nil {
		return nil, fmt.Errorf("unable to open config file (%s): %s", path, err)
	}

	if !json.Valid(rawData) {
		return nil, fmt.Errorf("invalid JSON found in file (%s)", path)
	}

	data := map[string]any{}
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config file (%s): %s", path, err)
	}

	return data, nil
}

func processConfigData(data map[string]any) error {
	for key, value := range data {
		valueJson, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("unable to marshal key (%s) data: %s", key, err)
		}

		ptr, registered := cfg.configPtrs[key]
		if !registered {
			cfg.configs[key] = valueJson
			continue
		}

		if err = json.Unmarshal(valueJson, ptr); err != nil {
			return fmt.Errorf("unable to unmarshal pointer for %s: %s", key, err)
		}
	}

	return nil
}

func loadEnvOverlay() error {
	if cfg.envFile == "" {
		return nil
	}

	_, statErr := os.Stat(cfg.envFile)
	cfg.envFileExisted = statErr == nil

	if !cfg.envFileExisted {
		return nil
	}

	envData, err := readJSONFile(cfg.envFile)
	if err != nil {
		return err
	}

	return processConfigData(envData)
}

func load() error {
	lock.Lock()
	defer lock.Unlock()

	baseData, err := readJSONFile(cfg.baseFile)
	if err != nil {
		return err
	}

	if err := processConfigData(baseData); err != nil {
		return err
	}

	if err := loadEnvOverlay(); err != nil {
		return err
	}

	for _, ptr := range cfg.configPtrs {
		initializeIfSupported(ptr)
	}

	cfg.lastLoad = time.Now()
	return nil
}

// DisableHotReload prevents the config file from being watched for changes.
// If called before LoadConfig, the watcher goroutine is never started.
// If called after, the existing watcher is stopped.
// Safe to call multiple times.
func DisableHotReload() {
	hotReloadDisabled.Store(true)
	hotReloadStopOnce.Do(func() {
		if hotReloadStop != nil {
			close(hotReloadStop)
		}
	})
}

func shouldReload() bool {
	lock.Lock()
	lastLoad := cfg.lastLoad
	envFileExisted := cfg.envFileExisted
	lock.Unlock()

	baseInfo, err := os.Stat(cfg.baseFile)
	if err != nil {
		log.Printf("config reload: unable to stat file (%s): %s", cfg.baseFile, err)
		return false
	}

	if baseInfo.ModTime().Sub(lastLoad) > 0 {
		return true
	}

	if cfg.envFile != "" {
		envInfo, err := os.Stat(cfg.envFile)
		envExists := err == nil

		if envExists && envInfo.ModTime().Sub(lastLoad) > 0 {
			return true
		}

		if envExists != envFileExisted {
			return true
		}
	}

	return false
}

func startHotReload() {
	if hotReloadDisabled.Load() {
		return
	}
	hotReloadStart.Do(func() {
		if hotReloadDisabled.Load() {
			return
		}
		hotReloadStop = make(chan struct{})
		ticker := time.NewTicker(time.Duration(checkInterval) * time.Second)
		go func() {
			for {
				select {
				case <-hotReloadStop:
					ticker.Stop()
					return
				case <-ticker.C:
					if shouldReload() {
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
// If the config struct implements the Initializer interface, Initialize() is called automatically
// after the first load and again on each hot reload.
//
// This function also sets up a check of the config file for any modifications. If changes are detected the config will be
// reloaded. Any errors encountered during a reload are logged and the previous configuration is retained.
// Hot reloading can be disabled by calling DisableHotReload.
func LoadConfig(name string, configStruct any) {
	cfg = getConfig()
	startHotReload()

	lock.Lock()
	defer lock.Unlock()

	cfgPtr, ok := cfg.configPtrs[name]
	if ok {
		configStruct = cfgPtr
	} else {
		configData, ok := cfg.configs[name]
		if !ok {
			panic(fmt.Errorf("key (%s) not found in config file (%s)", name, cfg.baseFile))
		}
		err := json.Unmarshal(configData, configStruct)
		if err != nil {
			panic(fmt.Errorf("unable to unmarshal (%s) to struct: %s", configData, err))
		}
		cfg.configPtrs[name] = configStruct
		initializeIfSupported(configStruct)
	}
}
