// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package engine

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"
	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/tracker"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/utils/settingspanel"
)

// Assuming Plugin type is properly defined in another file
type Plugin struct {
	API PluginAPI
}

// Interface definitions
type Engine interface {
	Availability
	Calendar
	EventResponder
	Subscriptions
	Users
	Welcomer
	Settings
	DailySummary
}

// Dependencies contains all API dependencies
type Dependencies struct {
	Logger            bot.Logger
	PluginAPI         PluginAPI
	Poster            bot.Poster
	Remote            remote.Remote
	Store             store.Store
	SettingsPanel     settingspanel.Panel
	IsAuthorizedAdmin func(string) (bool, error)
	Welcomer          Welcomer
	Tracker           tracker.Tracker
}

type PluginAPI interface {
	GetMattermostUser(mattermostUserID string) (*model.User, error)
	GetMattermostUserByUsername(mattermostUsername string) (*model.User, error)
	GetMattermostUserStatus(mattermostUserID string) (*model.Status, error)
	GetMattermostUserStatusesByIds(mattermostUserIDs []string) ([]*model.Status, error)
	IsSysAdmin(mattermostUserID string) (bool, error)
	UpdateMattermostUserStatus(mattermostUserID, status string) (*model.Status, error)
	UpdateMattermostUserCustomStatus(mattermostUserID string, customStatus *model.CustomStatus) *model.AppError
	RemoveMattermostUserCustomStatus(mattermostUserID string) *model.AppError
	GetPost(postID string) (*model.Post, error)
	CanLinkEventToChannel(channelID, userID string) bool
	SearchLinkableChannelForUser(teamID, mattermostUserID, search string) ([]*model.Channel, error)
	GetMattermostUserTeams(mattermostUserID string) ([]*model.Team, error)
	PublishWebsocketEvent(mattermostUserID, event string, payload map[string]any)
	GetUserPreferences(userID string) ([]*model.Preference, error)          // Added method for retrieving user preferences
	SendEphemeralPost(userID string, post *model.Post) (*model.Post, error) // Ensure this method exists
}

type Env struct {
	*config.Config
	*Dependencies
}

type mscalendar struct {
	Env

	actingUser *User
	client     remote.Client
	plugin     *Plugin
}

// copy returns a copy of the calendar engine
func (m *mscalendar) copy() *mscalendar {
	user := *m.actingUser
	client := m.client
	return &mscalendar{
		Env:        m.Env,
		actingUser: &user,
		client:     client,
		plugin:     m.plugin,
	}
}

func New(env Env, actingMattermostUserID string) Engine {
	return &mscalendar{
		Env:        env,
		actingUser: NewUser(actingMattermostUserID),
		plugin:     &Plugin{API: env.PluginAPI},
	}
}

func formatTime(t time.Time, format string) string {
	if format == "24-hour" {
		return t.Format("15:04")
	}
	return t.Format("03:04 PM")
}

func (m *mscalendar) getUserTimeFormat(userID string) (string, error) {
	user, err := m.plugin.API.GetMattermostUser(userID)
	if err != nil {
		return "", err
	}
	fmt.Println("User: ", user)

	// Replace with actual implementation to get user preferences
	preferences, err := m.plugin.API.GetUserPreferences(userID)
	if err != nil {
		return "", err
	}

	timeFormat := "12-hour" // Default value
	for _, pref := range preferences {
		if pref.Category == model.PreferenceCategoryDisplaySettings && pref.Name == model.PreferenceNameUseMilitaryTime {
			if pref.Value == "true" {
				timeFormat = "24-hour"
			}
			break
		}
	}

	return timeFormat, nil
}

func convertDateTime(dt *remote.DateTime) time.Time {
	// Implement this function based on the actual structure of remote.DateTime
	return time.Time{} // Placeholder
}

func (m *mscalendar) PostEventToUser(userID string, event *remote.Event) error {
	timeFormat, err := m.getUserTimeFormat(userID)
	if err != nil {
		return err
	}

	// Convert remote.DateTime to time.Time
	startTime := convertDateTime(event.Start)
	endTime := convertDateTime(event.End)

	// Using the timeFormat to format the event's time before posting it
	formattedStartTime := formatTime(startTime, timeFormat)
	formattedEndTime := formatTime(endTime, timeFormat)

	// Ensure that event.Title is available in remote.Event or use a default
	title := "Unknown Event Title" // Placeholder
	if event != nil && event.Subject != "" {
		title = event.Subject
	}

	// Posting the event to the user
	message := fmt.Sprintf("Event: %s from %s to %s", title, formattedStartTime, formattedEndTime)
	_, err = m.plugin.API.SendEphemeralPost(userID, &model.Post{
		Message: message,
	})
	return err
}
