// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/views"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/tz"
	"github.com/pkg/errors"
)

const dailySummaryTimeWindow = time.Minute * 2

var timeNowFunc = time.Now

type DailySummary interface {
	GetDailySummarySettingsForUser(user *User) (*store.DailySummaryUserSettings, error)
	SetDailySummaryPostTime(user *User, timeStr string) (*store.DailySummaryUserSettings, error)
	SetDailySummaryEnabled(user *User, enable bool) (*store.DailySummaryUserSettings, error)
	ProcessAllDailySummary() error
	GetDailySummary(user *User) (string, error)
}

func (m *mscalendar) GetDailySummarySettingsForUser(user *User) (*store.DailySummaryUserSettings, error) {
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

func (m *mscalendar) SetDailySummaryPostTime(user *User, timeStr string) (*store.DailySummaryUserSettings, error) {
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

	dsum, err := m.Store.LoadDailySummaryUserSettings(user.MattermostUserID)
	if err != nil {
		return nil, err
	}
	if dsum == nil {
		dsum = &store.DailySummaryUserSettings{MattermostUserID: user.MattermostUserID}
	}

	dsum.PostTime = timeStr
	dsum.Timezone = timezone

	err = m.Store.StoreDailySummaryUserSettings(dsum)
	if err != nil {
		return nil, err
	}
	return dsum, nil
}

func (m *mscalendar) SetDailySummaryEnabled(user *User, enable bool) (*store.DailySummaryUserSettings, error) {
	err := m.Filter(
		withClient,
		withUserExpanded(user),
	)
	if err != nil {
		return nil, err
	}

	dsum, err := m.Store.LoadDailySummaryUserSettings(user.MattermostUserID)
	if err != nil {
		return nil, err
	}
	if dsum == nil {
		dsum = &store.DailySummaryUserSettings{MattermostUserID: user.MattermostUserID}
	}
	dsum.Enable = enable

	err = m.Store.StoreDailySummaryUserSettings(dsum)
	if err != nil {
		return nil, err
	}
	return dsum, nil
}

func (m *mscalendar) ProcessAllDailySummary() error {
	err := m.Filter(
		withSuperuserClient,
	)
	if err != nil {
		return err
	}

	dsumIndex, err := m.Store.LoadDailySummaryIndex()
	if err != nil {
		return err
	}

	updatedPostTimes := map[string]string{}
	for _, dsum := range dsumIndex {
		shouldPost, err := shouldPostDailySummary(dsum)
		if err != nil {
			m.Logger.Errorf("Error posting daily summary for user %s: %s", dsum.MattermostUserID, err.Error())
			continue
		}
		if !shouldPost {
			continue
		}

		err = m.processDailySummary(dsum)
		if err != nil {
			m.Logger.Errorf("Error posting daily summary for user %s: %s", dsum.MattermostUserID, err.Error())
		}
		updatedPostTimes[dsum.MattermostUserID] = timeNowFunc().Format(time.RFC3339)
	}

	modErr := m.Store.ModifyDailySummaryIndex(func(dsumIndex store.DailySummaryIndex) (store.DailySummaryIndex, error) {
		for _, setting := range dsumIndex {
			postTime, ok := updatedPostTimes[setting.MattermostUserID]
			if ok {
				setting.LastPostTime = postTime
			}
		}
		return dsumIndex, nil
	})
	if modErr != nil {
		return modErr
	}

	return nil
}

func (m *mscalendar) GetDailySummary(user *User) (string, error) {
	return m.getDailySummary(user)
}

func (m *mscalendar) processDailySummary(dsum *store.DailySummaryUserSettings) error {
	user := NewUser(dsum.MattermostUserID)
	postStr, err := m.getDailySummary(user)
	if err != nil {
		return err
	}
	m.Poster.DM(user.MattermostUserID, postStr)
	return nil
}

func (m *mscalendar) getDailySummary(user *User) (string, error) {
	tz, err := m.GetTimezone(user)
	if err != nil {
		return "", err
	}

	calendarData, err := m.getTodayCalendar(user, tz)
	if err != nil {
		m.Poster.DM(user.MattermostUserID, "Failed to run daily summary job. %s", err.Error())
		return "", err
	}

	if len(calendarData) == 0 {
		m.Poster.DM(user.MattermostUserID, "You have no upcoming events today.")
		return "", nil
	}

	postStr, err := views.RenderCalendarView(calendarData, tz)
	if err != nil {
		return "", err
	}

	return postStr, nil
}

func shouldPostDailySummary(dsum *store.DailySummaryUserSettings) (bool, error) {
	if !dsum.Enable {
		return false, nil
	}

	lastPostStr := dsum.LastPostTime
	if lastPostStr != "" {
		lastPost, err := time.Parse(time.RFC3339, lastPostStr)
		if err != nil {
			return false, errors.New("Failed to parse last post time: " + lastPostStr)
		}
		since := timeNowFunc().Sub(lastPost)
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

	now := timeNowFunc().In(loc)
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
