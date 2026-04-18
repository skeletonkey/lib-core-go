package pushover

// Token information for the call to pushover API
//
//	Account: your user key
//	Application: application token
type token struct {
	Account     string `json:"account"`
	Application string `json:"application"`
}

// DisabledError is returned by Notify when pushover is disabled in the configuration.
// Callers can check for this error to distinguish a disabled notification from a real failure:
//
//	err := pushover.Notify(ctx, "hello")
//	if errors.As(err, &pushover.DisabledError{}) {
//	    // pushover is disabled, not an actual error
//	}
type DisabledError struct{}

func (e DisabledError) Error() string {
	return "pushover is disabled"
}

// Setting for use of the Pushover API:
//
//	 Enabled: bool to dis(en)able calling the API
//		URL: something like 'https://api.pushover.net/1'
//	 Token: see token struct info
type pushover struct {
	Token   token  `json:"token"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}
