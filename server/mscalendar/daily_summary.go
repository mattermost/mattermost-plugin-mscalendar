// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/tz"
	"github.com/pkg/errors"
)

const dailySummaryTimeWindow = time.Minute * 2

var timeNowFunc = time.Now

type DailySummary interface {
	GetDailySummarySettingsForUser(user *User) (*store.DailySummarySettings, error)
	SetDailySummaryPostTime(user *User, timeStr string) (*store.DailySummarySettings, error)
	SetDailySummaryEnabled(user *User, enable bool) (*store.DailySummarySettings, error)
	ProcessAllDailySummary(now time.Time) error
	GetDailySummaryForUser(user *User) (string, error)
}

func (m *mscalendar) GetDailySummarySettingsForUser(user *User) (*store.DailySummarySettings, error) {
	dsumIndex, err := m.Store.LoadDailySummaryIndex()
	if err != nil {
		return nil, err
	}

	for _, dsum := range dsumIndex {
		if dsum.MattermostUserID == user.MattermostUserID {
			return dsum, nil
		}
	}

	return nil, errors.New("No daily summary settings found")
}

func (m *mscalendar) SetDailySummaryPostTime(user *User, timeStr string) (*store.DailySummarySettings, error) {
	t, err := time.Parse(time.Kitchen, timeStr)
	if err != nil {
		return nil, errors.New("Invalid time value: " + timeStr)
	}

	if t.Minute()%int(dailySummaryJobInterval/time.Minute) != 0 {
		return nil, errors.Errorf("Time must be a multiple of %d minutes.", dailySummaryJobInterval/time.Minute)
	}

	timezone, err := m.GetTimezone(user)
	if err != nil {
		return nil, err
	}

	dsumIndex, err := m.Store.LoadDailySummaryIndex()
	if err != nil {
		return nil, err
	}

	var result *store.DailySummarySettings
	for _, dsum := range dsumIndex {
		if dsum.MattermostUserID == user.MattermostUserID {
			dsum.PostTime = timeStr
			dsum.Timezone = timezone
			result = dsum
			break
		}
	}
	if result == nil {
		err := m.Filter(withRemoteUser(user))
		if err != nil {
			return nil, err
		}
		result = &store.DailySummarySettings{
			MattermostUserID: user.MattermostUserID,
			RemoteID:         user.Remote.ID,
			PostTime:         timeStr,
			Timezone:         timezone,
		}
		dsumIndex = append(dsumIndex, result)
	}

	err = m.Store.SaveDailySummaryIndex(dsumIndex)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *mscalendar) SetDailySummaryEnabled(user *User, enable bool) (*store.DailySummarySettings, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	dsumIndex, err := m.Store.LoadDailySummaryIndex()
	if err != nil {
		return nil, err
	}

	var result *store.DailySummarySettings
	for _, dsum := range dsumIndex {
		if dsum.MattermostUserID == user.MattermostUserID {
			dsum.Enable = enable
			result = dsum
		}
	}
	if result == nil {
		err := m.Filter(withRemoteUser(user))
		if err != nil {
			return nil, err
		}
		result = &store.DailySummarySettings{
			MattermostUserID: user.MattermostUserID,
			RemoteID:         user.Remote.ID,
			Enable:           enable,
		}
		dsumIndex = append(dsumIndex, result)
	}

	err = m.Store.SaveDailySummaryIndex(dsumIndex)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *mscalendar) ProcessAllDailySummary(now time.Time) error {
	isAdmin, err := m.IsAuthorizedAdmin(m.actingUser.MattermostUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.Errorf("Non-admin user attempting ProcessAllDailySummary %s", m.actingUser.MattermostUserID)
	}

	dsumIndex, err := m.Store.LoadDailySummaryIndex()
	if err != nil {
		return err
	}
	if len(dsumIndex) == 0 {
		return nil
	}

	err = m.Filter(withSuperuserClient)
	if err != nil {
		return err
	}

	requests := []*remote.ViewCalendarParams{}
	for _, dsum := range dsumIndex {
		shouldPost, err := shouldPostDailySummary(dsum, now)
		if err != nil {
			m.Logger.Errorf("Error posting daily summary for user %s: %s", dsum.MattermostUserID, err.Error())
			continue
		}
		if !shouldPost {
			continue
		}

		start, end := getTodayHoursForTimezone(now, dsum.Timezone)
		req := &remote.ViewCalendarParams{
			RemoteID:  dsum.RemoteID,
			StartTime: start,
			EndTime:   end,
		}
		requests = append(requests, req)
	}

	responses, err := m.client.DoBatchViewCalendarRequests(requests)
	if err != nil {
		return err
	}

	mappedPostTimes := map[string]string{}
	byRemoteID := dsumIndex.ByRemoteID()
	for _, res := range responses {
		dsum := byRemoteID[res.RemoteID]
		if res.Error != nil {
			m.Logger.Errorf("Error rendering user %s calendar: %s %s", dsum.MattermostUserID, res.Error.Code, res.Error.Message)
		}
		postStr, err := views.RenderCalendarView(res.Events, dsum.Timezone)
		if err != nil {
			m.Logger.Errorf("Error rendering user %s calendar: %s", dsum.MattermostUserID, err.Error())
		}

		m.Poster.DM(dsum.MattermostUserID, postStr)
		mappedPostTimes[dsum.MattermostUserID] = time.Now().Format(time.RFC3339)
	}
	m.Logger.Infof("Processed daily summary for %d users", len(mappedPostTimes))

	// TODO atomic save index
	return nil
}

func (m *mscalendar) GetDailySummaryForUser(user *User) (string, error) {
	tz, err := m.GetTimezone(user)
	if err != nil {
		return "", err
	}

	calendarData, err := m.getTodayCalendarEvents(user, time.Now(), tz)
	if err != nil {
		return "Failed to get calendar events", err
	}

	return views.RenderCalendarView(calendarData, tz)
}

func shouldPostDailySummary(dsum *store.DailySummarySettings, now time.Time) (bool, error) {
	if !dsum.Enable {
		return false, nil
	}

	lastPostStr := dsum.LastPostTime
	if lastPostStr != "" {
		lastPost, err := time.Parse(time.RFC3339, lastPostStr)
		if err != nil {
			return false, errors.New("Failed to parse last post time: " + lastPostStr)
		}
		since := now.Sub(lastPost)
		if since < dailySummaryTimeWindow {
			return false, nil
		}
	}

	timezone := tz.Go(dsum.Timezone)
	if timezone == "" {
		return false, errors.New("Invalid timezone")
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return false, err
	}
	t, err := time.ParseInLocation(time.Kitchen, dsum.PostTime, loc)
	if err != nil {
		return false, err
	}

	now = now.In(loc)
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false, nil
	}

	t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
	diff := now.Sub((t))
	if diff >= 0 {
		return diff < dailySummaryTimeWindow, nil
	}
	return -diff < dailySummaryTimeWindow, nil
}

func getTodayHoursForTimezone(now time.Time, timezone string) (start, end time.Time) {
	t := remote.NewDateTime(now.UTC(), "UTC").In(timezone).Time()
	start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end = start.Add(24 * time.Hour)
	return start, end
}
