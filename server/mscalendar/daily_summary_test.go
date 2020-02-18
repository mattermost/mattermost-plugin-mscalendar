package mscalendar

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/stretchr/testify/require"
)

func TestShouldPostDailySummary(t *testing.T) {
	loc, err := time.LoadLocation("EST")
	require.Nil(t, err)

	hour, minute := 9, 0 // Time is "9:00AM"
	moment := makeTime(hour, minute, loc)
	c := clock.NewMock()
	c.Set(moment)
	timeNowFunc = c.Now

	enabled := false
	timeStr := "8:00AM"
	timezoneStr := "Eastern Standard Time"
	dsum := &store.DailySummarySettings{
		Enable:   enabled,
		PostTime: timeStr,
		Timezone: timezoneStr,
	}

	shouldRun, err := shouldPostDailySummary(dsum)
	require.Nil(t, err)
	require.False(t, shouldRun)

	dsum.Enable = true
	shouldRun, err = shouldPostDailySummary(dsum)
	require.Nil(t, err)
	require.False(t, shouldRun)

	dsum.PostTime = "9:00AM"
	dsum.Timezone = "Eastern Standard Time"
	shouldRun, err = shouldPostDailySummary(dsum)
	require.Nil(t, err)
	require.True(t, shouldRun)

	dsum.PostTime = "9:00AM"
	dsum.Timezone = "Mountain Standard Time"
	shouldRun, err = shouldPostDailySummary(dsum)
	require.Nil(t, err)
	require.False(t, shouldRun)

	dsum.PostTime = "7:00AM"
	dsum.Timezone = "Mountain Standard Time"
	shouldRun, err = shouldPostDailySummary(dsum)
	require.Nil(t, err)
	require.True(t, shouldRun)

	dsum.PostTime = "7:20FM" // Invalid time
	dsum.Timezone = "Mountain Standard Time"
	shouldRun, err = shouldPostDailySummary(dsum)
	require.NotNil(t, err)
	require.False(t, shouldRun)

	dsum.PostTime = "7:00AM"
	dsum.Timezone = "Moon Time" // Invalid timezone
	shouldRun, err = shouldPostDailySummary(dsum)
	require.NotNil(t, err)
	require.False(t, shouldRun)
}

func makeTime(hour, minute int, loc *time.Location) time.Time {
	return time.Date(2020, 2, 12, hour, minute, 0, 0, loc)
}
