package mscalendar

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/stretchr/testify/require"
)

func TestShouldPostDailySummary(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		postTime    string
		timeZone    string
		shouldRun   bool
		shouldError bool
	}{
		{
			name:        "Disabled",
			enabled:     false,
			postTime:    "9:00AM",
			timeZone:    "Eastern Standard Time",
			shouldRun:   false,
			shouldError: false,
		},
		{
			name:        "Same timezone, wrong time",
			enabled:     true,
			postTime:    "8:00AM",
			timeZone:    "Eastern Standard Time",
			shouldRun:   false,
			shouldError: false,
		},
		{
			name:        "Same timezone, right time",
			enabled:     true,
			postTime:    "9:00AM",
			timeZone:    "Eastern Standard Time",
			shouldRun:   true,
			shouldError: false,
		},
		{
			name:        "Different timezone, wrong time",
			enabled:     true,
			postTime:    "9:00AM",
			timeZone:    "Mountain Standard Time",
			shouldRun:   false,
			shouldError: false,
		},
		{
			name:        "Different timezone, right time",
			enabled:     true,
			postTime:    "7:00AM",
			timeZone:    "Mountain Standard Time",
			shouldRun:   true,
			shouldError: false,
		},
		{
			name:        "Nepal timezone, wrong time",
			enabled:     true,
			postTime:    "7:00AM",
			timeZone:    "Nepal Standard Time",
			shouldRun:   false,
			shouldError: false,
		},
		{
			name:        "Nepal timezone, right time",
			enabled:     true,
			postTime:    "7:45PM",
			timeZone:    "Nepal Standard Time",
			shouldRun:   true,
			shouldError: false,
		},
		{
			enabled:     true,
			postTime:    "7:20FM", // Invalid time
			timeZone:    "Mountain Standard Time",
			shouldRun:   false,
			shouldError: true,
		},
		{
			enabled:     true,
			postTime:    "7:00AM",
			timeZone:    "Moon Time",
			shouldRun:   false,
			shouldError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := time.LoadLocation("EST")
			require.Nil(t, err)

			hour, minute := 9, 0 // Time is "9:00AM"
			moment := makeTime(hour, minute, loc)
			c := clock.NewMock()
			c.Set(moment)
			timeNowFunc = c.Now

			dsum := &store.DailySummarySettings{
				Enable:   tc.enabled,
				PostTime: tc.postTime,
				Timezone: tc.timeZone,
			}

			shouldRun, err := shouldPostDailySummary(dsum)
			require.Equal(t, tc.shouldRun, shouldRun)
			if tc.shouldError {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func makeTime(hour, minute int, loc *time.Location) time.Time {
	return time.Date(2020, 2, 12, hour, minute, 0, 0, loc)
}
