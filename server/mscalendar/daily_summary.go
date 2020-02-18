// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"time"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/tz"
	"github.com/pkg/errors"
)

const dailySummaryTimeWindow = time.Minute * 2

var timeNowFunc = time.Now

type DailySummary interface {
	GetDailySummarySettingsForUser(user *User) (*store.DailySummarySettings, error)
	SetDailySummaryPostTime(user *User, timeStr string) (*store.DailySummarySettings, error)
	SetDailySummaryEnabled(user *User, enable bool) (*store.DailySummarySettings, error)
	DailySummaryAll() error
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
		return nil, errors.New("Invalid time value")
	}

	if t.Minute() != 0 && t.Minute() != 30 {
		return nil, errors.New("Time must be a multiple of 30 minutes.")
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
		result = &store.DailySummarySettings{
			MattermostUserID: user.MattermostUserID,
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
		result = &store.DailySummarySettings{
			MattermostUserID: user.MattermostUserID,
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

func (m *mscalendar) DailySummaryAll() error {
	dsumIndex, err := m.Store.LoadDailySummaryIndex()
	if err != nil {
		return err
	}

	for _, dsum := range dsumIndex {
		shouldPost, err := shouldPostDailySummary(dsum)
		if err != nil {
			return err
		}
		if !shouldPost {
			continue
		}

		err = m.postDailySummary(dsum.MattermostUserID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mscalendar) postDailySummary(mattermostUserID string) error {
	user := NewUser(mattermostUserID)
	calendarData, err := m.viewTodayCalendar(user)
	if err != nil {
		m.Poster.DM(mattermostUserID, "Failed to run daily summary job. %s", err.Error())
		return err
	}
	if len(calendarData) == 0 {
		m.Poster.DM(mattermostUserID, "You have no upcoming events today.")
	} else {
		m.Poster.DM(mattermostUserID, utils.JSONBlock(calendarData))
	}
	return nil
}

func shouldPostDailySummary(dsum *store.DailySummarySettings) (bool, error) {
	if !dsum.Enable {
		return false, nil
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
