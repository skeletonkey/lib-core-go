# CHANGELOG

## 1.0.0

- BREAKING CHANGE: `PROJECT_CONFIG_FILE` now expects a filename (not a full path); use `PROJECT_CONFIG_DIR` for directory
- BREAKING CHANGE: pushover's Notify method now requires a context when being called
- BREAKING CHANGE: pushover's Notify returns DisabledError when pushover is disabled; logging removed from library
- config: add environment-based config overlays via `PROJECT_CONFIG_DIR` and `PROJECT_CONFIG_ENV_VAR`
- config: support multiple search directories (colon-separated `PROJECT_CONFIG_DIR`)
- logger: add HandleErr method
- upgrade Go to 1.26.2
- upgrade dependencies to latest versions
- linting and other cleanup

## 0.3.0

- constant RACHIO_CONFIG_FILE changed to PROJECT_CONFIG_FILE - note this is somewhat of a breaking change
- update Go to 1.22
- add pushover notifier
