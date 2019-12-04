// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package remote

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateTime_Time(t *testing.T) {
	for _, tc := range []struct {
		DateTime
		expectedUTC    string
		expectedString string
	}{
		{
			DateTime:       DateTime{"2019-11-19T01:00:00.0000005", ""},
			expectedUTC:    "2019-11-19T01:00:00.0000005Z",
			expectedString: "2019-11-19T01:00:00Z",
		},
		{
			DateTime:       DateTime{"2019-11-19T01:00:00.9999999", "Asia/Hong_Kong"},
			expectedUTC:    "2019-11-18T17:00:00.9999999Z",
			expectedString: "2019-11-19T01:00:00+08:00",
		},
		{
			DateTime:       DateTime{"2019-11-19T14:00:00.0000005", "MST"},
			expectedUTC:    "2019-11-19T21:00:00.0000005Z",
			expectedString: "2019-11-19T14:00:00-07:00",
		},
	} {
		t.Run(tc.TimeZone, func(t *testing.T) {
			require.Equal(t, tc.expectedUTC, tc.DateTime.Time().UTC().Format(time.RFC3339Nano))
			require.Equal(t, tc.expectedString, tc.DateTime.String())
		})
	}
}
