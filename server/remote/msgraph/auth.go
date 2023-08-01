package msgraph

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/pkg/errors"
	msgraph "github.com/yaegashi/msgraph.go/v1.0"
)

func (c *client) RevokeSession(session remote.Session) error {
	_, err := c.rbuilder.Users().ID("").
		RevokeSignInSessions(&msgraph.UserRevokeSignInSessionsRequestParameter{}).
		Request().Post(c.ctx)
	if err != nil {
		return errors.Wrap(err, "msgraph RevokeToken")
	}

	return nil
}
