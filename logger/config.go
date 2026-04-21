package logger

import "github.com/skeletonkey/lib-core-go/config"

//nolint:gochecknoglobals // cfg holds the configuration data, which should only be retrieved and not changed
var cfg *Logger

func getConfig() *Logger {
	config.LoadConfig("logger", &cfg)
	return cfg
}
