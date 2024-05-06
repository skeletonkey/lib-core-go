// Send message using pushover.net
package pushover

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/skeletonkey/lib-core-go/logger"
)

// Notify sends `msg` using the Pushover API
func Notify(msg string) {
	config := getConfig()
	log := logger.Get()
	requestUrl := fmt.Sprintf("%s/messages.json?token=%s&user=%s&message=%s",
		config.URL, config.Token.Application, config.Token.Account, url.QueryEscape(msg))
	log.Debug().Str("URL", requestUrl).Msg("notification URL")
	if !config.Enabled {
		log.Info().Msg("Pushover is disabled")
	} else {
		res, err := http.Post(requestUrl, "application/json", nil)
		if err != nil {
			log.Panic().Err(err).Str("URL", requestUrl).Msg("unable to post to url")
		}
		body, err := io.ReadAll(res.Body)
		defer func() {
			err := res.Body.Close()
			if err != nil {
				log.Error().
					Err(err).
					Msg("unable to close response body")
			}
		}()
		if err != nil {
			log.Panic().Err(err).Interface("response", res).Msg("unable to read response body")
		}
		if res.StatusCode != 200 {
			log.Panic().Int("Status Code", res.StatusCode).Bytes("response body", body).Msg("non-200 response received")
		}

		log.Info().Bytes("response body", body).Msg("pushover response")
	}
}