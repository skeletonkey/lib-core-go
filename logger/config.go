package logger

import "github.com/skeletonkey/lib-core-go/config"

var cfg *Logger

func getConfig() *Logger {
	config.LoadConfig("logger", &cfg)
	return cfg
}
