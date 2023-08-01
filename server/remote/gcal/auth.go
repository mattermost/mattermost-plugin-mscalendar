package gcal

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
)

func (c *client) RevokeSession(session remote.Session) error {
	resp, err := http.PostForm("https://oauth2.googleapis.com/revoke", url.Values{
		"token": []string{session.AccessToken},
	})
	if err != nil {
		return errors.Wrap(err, "error revoking user accesss token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body []byte
		var err error

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "error reading revoke response body")
		}

		c.Logger.With(bot.LogContext{
			"response": string(body),
		}).Errorf("error revoking token")
		return fmt.Errorf("unsuccessful revoke request")
	}

	return nil
}
