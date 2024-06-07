# lib-core-go

Core libraries for running a service

## config

Configuration library that reads from a JSON file and supports 'hot reloading.'

Location of the project's config JSON file is currently in the environmental variable `PROJECT_CONFIG_FILE`.

### JSON config file

The config file should be a JSON object with the top level keys being used as identifiers to the values (also JSON objects) which will be mapped to structs.

#### Example

JSON:

```json
{
  "app": {
    "name": "testing",
    "version": 1,
    "enabled": true
  }
}
```

Go Code:

```
type app struct {
    Name    string `json:"name"`
    Ver     int    `json:"version"`
    Enabled bool `json:"enabled"`
}

func main() {
  var appData app
  config.LoadConfig("app", &appData)
}
```

### Caveat Emptor

At this time, 'secrets' are written directly into the config file.
It is _HIGHLY_ recommended that you DO NOT check it into source control.
This will hopefully be addressed in the future.

Configuration allows for 'hot reloading.'
If you change the configuration, it should be live within 15 seconds (currently this is not configurable).

Hot Reloading things to know:

* if the config file is modified, changes will be reloaded in 15 seconds
  * this is not editable
  * you cannot disable this behavior
* changing configs for objects that utilize an Initializer doesn't have the desired effect
  * recommend not changing these settings
    * if changes are made, restart the service
    * bug report (entry in the TODO file) has been submitted
* if the config is messed up (like breaking the JSON), it will cause the program to panic
  * recommended not to use an editor that auto saves without linting
  * bug report has been submitted

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

```
  "pushover": {
    "enabled": true,
    "url": "https://api.pushover.net/1",
    "token": {
      "account": "ACCOUNT TOKEN GOES HERE (YOUR USER KEY)",
      "application": "APPLICATION TOKEN GOES HERE"
    }
  }
```
