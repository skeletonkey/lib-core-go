# lib-core-go

Core libraries for running a service

## config

Configuration library that reads from JSON files and supports hot reloading with environment-based overlays.

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `PROJECT_CONFIG_FILE` | Yes | — | Base config filename (e.g., `config.json`) |
| `PROJECT_CONFIG_DIR` | No | `.` | Colon-separated list of directories to search for config files. First directory containing the base file wins. |
| `PROJECT_CONFIG_ENV_VAR` | No | — | Name of the env var holding the environment name. Used to find the overlay file. |
| `PROJECT_CONFIG_CHECK_INTERVAL` | No | `15` | Hot reload check interval in seconds. Must be a positive integer. |

### Environment Overlays

When `PROJECT_CONFIG_ENV_VAR` is set, the system reads the env var it names to determine the environment. An overlay file named `{base}_{env}.{ext}` is loaded on top of the base config.

For example, with `PROJECT_CONFIG_FILE=config.json`, `PROJECT_CONFIG_ENV_VAR=DEPLOY_ENV`, and `DEPLOY_ENV=prod`:
1. Searches for `config.json` in each directory listed in `PROJECT_CONFIG_DIR`
2. Loads `config.json` as the base configuration
3. If `config_prod.json` exists in the same directory, overlays it on top

The overlay only needs to specify fields that differ from the base. Fields not present in the overlay retain their base values. This means `config.json` can be safely checked into source control with non-sensitive defaults, while `config_prod.json` holds environment-specific overrides.

### JSON Config File

The config file should be a JSON object with the top level keys being used as identifiers to the values (also JSON objects) which will be mapped to structs.

#### Example

Base config (`config.json`):

```json
{
  "app": {
    "name": "testing",
    "version": 1,
    "enabled": true
  }
}
```

Production overlay (`config_prod.json`):

```json
{
  "app": {
    "enabled": false
  }
}
```

Result: `app` struct has `name: "testing"`, `version: 1`, `enabled: false`.

Go Code:

```go
type app struct {
    Name    string `json:"name"`
    Ver     int    `json:"version"`
    Enabled bool   `json:"enabled"`
}

func main() {
  var appData app
  config.LoadConfig("app", &appData)
}
```

### Initializer Interface

If your config struct implements `config.Initializer` (an `Initialize()` method), it will be called automatically after the config is loaded and again on each hot reload.

### Hot Reloading

Config files are checked for modifications every 15 seconds. If changes are detected, the config is reloaded. Errors during reload are logged and the previous configuration is retained.

Hot reloading can be disabled by calling `config.DisableHotReload()` before or after `LoadConfig`.

Both the base file and environment overlay file are watched for changes.

## logger

Package logger provides a convenience method for setting up rs/zerolog and returning a logger.

See rs/zerlog documentation for how to properly log.

To use properly ensure that the following exists in the application's config:

```json
{
  "logger": {
    "log_file": "/var/log/my_application.log",
    "log_level": "debug",
    "to_file": true,
    "to_stderr": false,
    "to_stdout": false
  }
}
```

## notify

### pushover

Use pushover.net's API to send messages to your phone's Pushover App.

This relies on config and logger modules.

```json
  "pushover": {
    "enabled": true,
    "url": "https://api.pushover.net/1",
    "token": {
      "account": "ACCOUNT TOKEN GOES HERE (YOUR USER KEY)",
      "application": "APPLICATION TOKEN GOES HERE"
    }
  }
```
