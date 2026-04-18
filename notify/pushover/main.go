// Send message using pushover.net
package pushover

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/skeletonkey/lib-core-go/logger"
)

// Notify sends `msg` using the Pushover API
// Not timeout is set for the context - the caller should set one if desired
func Notify(ctx context.Context, msg string) (err error) {
	config := getConfig()
	log := logger.Get()
	requestUrl := fmt.Sprintf("%s/messages.json?token=%s&user=%s&message=%s",
		config.URL, config.Token.Application, config.Token.Account, url.QueryEscape(msg))
	log.Trace().Str("URL", requestUrl).Msg("notification URL")

	if !config.Enabled {
		log.Info().Msg("Pushover is disabled")
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", requestUrl, nil)
	if err != nil {
		log.Debug().Err(err).Str("URL", requestUrl).Msg("unable to create request")
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Debug().Err(err).Str("URL", requestUrl).Msg("unable to post to url")
		return err
	}
	defer func() { logger.HandleErr(res.Body.Close(), "unable to close response body") }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Debug().Err(err).Interface("response", res).Msg("unable to read response body")
		return err
	}
	if res.StatusCode != 200 {
		log.Debug().Int("Status Code", res.StatusCode).Bytes("response body", body).Msg("non-200 response received")
		return fmt.Errorf("non-200 response from pushover: %d", res.StatusCode)
	}

	log.Trace().Bytes("response body", body).Msg("pushover response")

	return nil
}
