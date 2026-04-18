// Send message using pushover.net
package pushover

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Notify sends `msg` using the Pushover API
// Not timeout is set for the context - the caller should set one if desired
func Notify(ctx context.Context, msg string) (err error) {
	config := getConfig()
	requestUrl := fmt.Sprintf("%s/messages.json?token=%s&user=%s&message=%s",
		config.URL, config.Token.Application, config.Token.Account, url.QueryEscape(msg))

	if !config.Enabled {
		return ErrDisabled{}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", requestUrl, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("non-200 response from pushover: %d", res.StatusCode)
	}

	return nil
}
