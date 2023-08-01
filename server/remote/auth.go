package remote

// Session represents an user authenticated with the plugin and the minimal data we need to revoke
// the sessions from our side.
type Session struct {
	RemoteID    string
	AccessToken string
}
