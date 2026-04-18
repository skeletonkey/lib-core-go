package pushover

import "github.com/skeletonkey/lib-core-go/config"

//nolint:gochecknoglobals // cfg holds the configuration data, which should only be retrieved and not changed
var cfg *pushover

func getConfig() *pushover {
	config.LoadConfig("pushover", &cfg)
	return cfg
}
