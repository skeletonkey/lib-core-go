package pushover

import "github.com/skeletonkey/lib-core-go/config"

var cfg *pushover

func getConfig() *pushover {
	config.LoadConfig("pushover", &cfg)
	return cfg
}
