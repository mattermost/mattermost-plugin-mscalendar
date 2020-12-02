// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/tz"
)

const dailySummaryTimeWindow = time.Minute * 2

// Run daily summary job every 15 minutes
const DailySummaryJobInterval = 15 * time.Minute

type DailySummary interface {
	GetDailySummaryForUser(user *User) (string, error)
	GetDailySummarySettingsForUser(user *User) (*store.DailySummaryUserSettings, error)
	SetDailySummaryPostTime(user *User, timeStr string) (*store.DailySummaryUserSettings, error)
	SetDailySummaryEnabled(user *User, enable bool) (*store.DailySummaryUserSettings, error)
	ProcessAllDailySummary(now time.Time) error
}

func (m *mscalendar) GetDailySummarySettingsForUser(user *User) (*store.DailySummaryUserSettings, error) {
	err := m.Filter(withUserExpanded(user))
	if err != nil {
		return nil, err
	}

	return user.Settings.DailySummary, nil
}

func (m *mscalendar) SetDailySummaryPostTime(user *User, timeStr string) (*store.DailySummaryUserSettings, error) {
	err := m.Filter(withUserExpanded(user))
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(time.Kitchen, timeStr)
	if err != nil {
		return nil, errors.New("Invalid time value: " + timeStr)
	}

	if t.Minute()%int(DailySummaryJobInterval/time.Minute) != 0 {
		return nil, fmt.Errorf("time must be a multiple of %d minutes", DailySummaryJobInterval/time.Minute)
	}

	timezone, err := m.GetTimezone(user)
	if err != nil {
		return nil, err
	}

	if user.Settings.DailySummary == nil {
		user.Settings.DailySummary = store.DefaultDailySummaryUserSettings()
	}

	dsum := user.Settings.DailySummary
	dsum.PostTime = timeStr
	dsum.Timezone = timezone

	err = m.Store.StoreUser(user.User)
	if err != nil {
		return nil, err
	}
	return dsum, nil
}

func (m *mscalendar) SetDailySummaryEnabled(user *User, enable bool) (*store.DailySummaryUserSettings, error) {
	err := m.Filter(withUserExpanded(user))
	if err != nil {
		return nil, err
	}

	if user.Settings.DailySummary == nil {
		user.Settings.DailySummary = store.DefaultDailySummaryUserSettings()
	}

	dsum := user.Settings.DailySummary
	dsum.Enable = enable

	err = m.Store.StoreUser(user.User)
	if err != nil {
		return nil, err
	}
	return dsum, nil
}

func (m *mscalendar) ProcessAllDailySummary(now time.Time) error {
	userIndex, err := m.Store.LoadUserIndex()
	if err != nil {
		return err
	}
	if len(userIndex) == 0 {
		return nil
	}

	err = m.Filter(withSuperuserClient)
	if err != nil {
		return err
	}

	requests := []*remote.ViewCalendarParams{}
	byRemoteID := map[string]*store.User{}
	for _, user := range userIndex {
		storeUser, storeErr := m.Store.LoadUser(user.MattermostUserID)
		if storeErr != nil {
			m.Logger.Warnf("Error loading user %s for daily summary. err=%v", user.MattermostUserID, storeErr)
			continue
		}
		byRemoteID[storeUser.Remote.ID] = storeUser

		dsum := storeUser.Settings.DailySummary
		if dsum == nil {
			continue
		}

		shouldPost, shouldPostErr := shouldPostDailySummary(dsum, now)
		if shouldPostErr != nil {
			m.Logger.Warnf("Error posting daily summary for user %s. err=%v", user.MattermostUserID, shouldPostErr)
			continue
		}
		if !shouldPost {
			continue
		}

		start, end := getTodayHoursForTimezone(now, dsum.Timezone)
		req := &remote.ViewCalendarParams{
			RemoteUserID: storeUser.Remote.ID,
			StartTime:    start,
			EndTime:      end,
		}
		requests = append(requests, req)
	}

	responses, err := m.client.DoBatchViewCalendarRequests(requests)
	if err != nil {
		return err
	}

	for _, res := range responses {
		user := byRemoteID[res.RemoteUserID]
		if res.Error != nil {
			m.Logger.Warnf("Error rendering user %s calendar. err=%s %s", user.MattermostUserID, res.Error.Code, res.Error.Message)
		}
		dsum := user.Settings.DailySummary
		if dsum == nil {
			// Should never reach this point
			continue
		}
		postStr, err := views.RenderCalendarView(res.Events, dsum.Timezone)
		if err != nil {
			m.Logger.Warnf("Error rendering user %s calendar. err=%v", user.MattermostUserID, err)
		}

		m.Poster.DM(user.MattermostUserID, postStr)
		m.Dependencies.Tracker.TrackDailySummarySent(user.MattermostUserID)
		dsum.LastPostTime = time.Now().Format(time.RFC3339)
		err = m.Store.StoreUser(user)
		if err != nil {
			m.Logger.Warnf("Error storing daily summary LastPostTime for user %s. err=%v", user.MattermostUserID, err)
		}
	}

	m.Logger.Infof("Processed daily summary for %d users", len(responses))
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

func shouldPostDailySummary(dsum *store.DailySummaryUserSettings, now time.Time) (bool, error) {
	if dsum == nil || !dsum.Enable {
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
		return false, errors.New("invalid timezone")
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
