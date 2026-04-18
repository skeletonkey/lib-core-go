// Send message using pushover.net
package pushover

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	if res.StatusCode/100 != 2 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("pushover API error (%d): unable to read response body", res.StatusCode)
		}
		var apiResp struct {
			Errors []string `json:"errors"`
		}
		if err := json.Unmarshal(body, &apiResp); err == nil && len(apiResp.Errors) > 0 {
			return fmt.Errorf("pushover API error (%d): %s", res.StatusCode, strings.Join(apiResp.Errors, "; "))
		}
		return fmt.Errorf("pushover API error (%d): unknown error", res.StatusCode)
	}

	return nil
}
