// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package msgraph

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost-plugin-msoffice/server/remote"
	"github.com/stretchr/testify/require"
)

func TestPrepareGetScheduleRequests(t *testing.T) {
	for name, tc := range map[string]struct {
		schedules     []string
		runAssertions func(t *testing.T, out []*singleRequest)
	}{
		"5 emails": {
			schedules: []string{"a", "b", "c", "d", "e"},
			runAssertions: func(t *testing.T, out []*singleRequest) {
				sched1 := []string{"a", "b", "c", "d", "e"}

				require.Equal(t, 1, len(out))
				require.Equal(t, "0", out[0].ID)
				body := out[0].Body.(*getScheduleRequest)
				require.Equal(t, 5, len(body.Schedules))
				require.Equal(t, sched1, body.Schedules)
			},
		},
		"20 emails": {
			schedules: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t"},
			runAssertions: func(t *testing.T, out []*singleRequest) {
				sched1 := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t"}

				require.Equal(t, 1, len(out))
				require.Equal(t, "0", out[0].ID)
				body := out[0].Body.(*getScheduleRequest)
				require.Equal(t, 20, len(body.Schedules))
				require.Equal(t, sched1, body.Schedules)
			},
		},
		"26 emails": {
			schedules: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
			runAssertions: func(t *testing.T, out []*singleRequest) {
				sched1 := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t"}
				sched2 := []string{"u", "v", "w", "x", "y", "z"}

				require.Equal(t, 2, len(out))
				require.Equal(t, "0", out[0].ID)
				body := out[0].Body.(*getScheduleRequest)
				require.Equal(t, 20, len(body.Schedules))
				require.Equal(t, sched1, body.Schedules)

				require.Equal(t, "1", out[1].ID)
				body = out[1].Body.(*getScheduleRequest)
				require.Equal(t, 6, len(body.Schedules))
				require.Equal(t, sched2, body.Schedules)
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			userID := "xyz"
			start := time.Now()
			end := time.Now().Add(20)
			params := &getScheduleRequest{
				StartTime:                remote.NewDateTime(start),
				EndTime:                  remote.NewDateTime(end),
				AvailabilityViewInterval: 15,
			}

			out := prepareGetScheduleRequests(userID, tc.schedules, params)
			require.Equal(t, "/Users/xyz/calendar/getSchedule", out[0].URL)
			require.Equal(t, "POST", out[0].Method)
			require.Equal(t, 1, len(out[0].Headers))
			require.Equal(t, "application/json", out[0].Headers["Content-Type"])

			body := out[0].Body.(*getScheduleRequest)
			require.Equal(t, params.StartTime.String(), body.StartTime.String())
			require.Equal(t, params.EndTime.String(), body.EndTime.String())
			require.Equal(t, 15, body.AvailabilityViewInterval)

			tc.runAssertions(t, out)
		})
	}
}
