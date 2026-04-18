# 1.0.0

- BREAKING CHANGE: pushover's Notify method now requires a context when being called
- BREAKING CHANGE: pushover's Notify returns DisabledError when pushover is disabled; logging removed from library
- logger: add HandleErr method
- upgrade Go to 1.26.2
- upgrade dependencies to latest versions
- linting and other cleanup

# 0.3.0

- constant RACHIO_CONFIG_FILE changed to PROJECT_CONFIG_FILE - note this is somewhat of a breaking change
- update Go to 1.22
- add pushover notifier
