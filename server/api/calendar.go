// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
)

func (api *api) ViewCalendar(from, to time.Time) ([]*remote.Event, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	// api.Config.Poster.Ephemeral()
	// api.Config.Config.BotConfig.AdminUserIDs
	// api.Config.Config.BotConfig
	// api.Config.Config
	return client.GetUserDefaultCalendarView(api.user.Remote.ID, from, to)
}

func (api *api) CreateCalendar(calendar *remote.Calendar) (*remote.Calendar, error) {

	// api.Config.Poster.Ephemeral()
	// conf := Plugin.getConfig()

	// conf := plugin.GetConfig()
	// plugin.API.GetUser()
	// p := plugin.NewWithConfig(conf)
	// p.API.GetUser("")

	// p.plugin..Plugin.API.GetUser("")
	// conf2 := p..getConfig()
	// p := Plugin.NewWithConfig(conf)
	// user, err := p.NewWithConfig(conf2).MattermostPlugin.API.GetUser("junk")
	// p.API.GetUser()
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.CreateCalendar(calendar)
}

func (api *api) CreateEvent(calendarEvent *remote.Event) (*remote.Event, error) {

	// var junk plugin.api.newAPIConfig()
	// conf := Plugin.getConfig()
	// p := Plugin.NewWithConfig(conf)
	// p

	value := api.BotConfig
	fmt.Printf("value = %+v\n", value)

	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.CreateEvent(calendarEvent)
}

func (api *api) DeleteCalendar(calendarID string) error {
	client, err := api.MakeClient()
	if err != nil {
		return err
	}

	return client.DeleteCalendar(calendarID)
}

func (api *api) FindMeetingTimes(meetingParams *remote.FindMeetingTimesParameters) (*remote.MeetingTimeSuggestionResults, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	return client.FindMeetingTimes(meetingParams)
}

func (api *api) GetUserCalendars(userID string) ([]*remote.Calendar, error) {
	client, err := api.MakeClient()
	if err != nil {
		return nil, err
	}

	//DEBUG just use me as the user, for now
	me, err := client.GetMe()
	return client.GetUserCalendars(me.ID)
}
