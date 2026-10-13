// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package views

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/calendar/remote"

	"github.com/stretchr/testify/require"
)

func TestMarkdownToHTMLEntities(t *testing.T) {
	for _, testCase := range []struct {
		description    string
		inputstring    string
		expectedOutput string
	}{
		{
			description:    "with asterisk",
			inputstring:    "**bold text**",
			expectedOutput: "&#42;&#42;bold text&#42;&#42;",
		},
		{
			description:    "normal string",
			inputstring:    "normal string",
			expectedOutput: "normal string",
		},
		{
			description:    "with braces",
			inputstring:    "[square](round)",
			expectedOutput: "&#91;square&#93;&#40;round&#41;",
		},
		{
			description:    "with underscore",
			inputstring:    "text_test",
			expectedOutput: "text&#95;test",
		},
		{
			description:    "withbacktick",
			inputstring:    "`test`",
			expectedOutput: "&#96;test&#96;",
		},
		{
			description:    "with greater and less than",
			inputstring:    "<test>",
			expectedOutput: "&#60;test&#62;",
		},
		{
			description:    "with backslash",
			inputstring:    "test \\ text",
			expectedOutput: "test &#92; text",
		},
		{
			description:    "URL 1",
			inputstring:    "www.example.com",
			expectedOutput: "www&#46;example&#46;com",
		},
		{
			description:    "URL 2",
			inputstring:    "https://example.com",
			expectedOutput: "https&#58;&#47;&#47;example&#46;com",
		},
		{
			description:    "strike through",
			inputstring:    "~~strike~~",
			expectedOutput: "&#126;&#126;strike&#126;&#126;",
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			res := MarkdownToHTMLEntities(testCase.inputstring)
			require.EqualValues(t, testCase.expectedOutput, res)
		})
	}
}

func TestEventTimeFormat(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Madrid")
	require.NoError(t, err)

	startTime := time.Date(2023, 8, 9, 18, 0, 0, 0, loc)
	endTime := time.Date(2023, 8, 9, 18, 15, 0, 0, loc)

	event := &remote.Event{
		Subject: "Team Sync",
		Start:   remote.NewDateTime(startTime, "Europe/Madrid"),
		End:     remote.NewDateTime(endTime, "Europe/Madrid"),
	}

	opt := ShowTimezoneOption("Europe/Madrid")
	attachment := &model.MessageAttachment{}
	opt.Apply(*event, attachment)

	require.Equal(t, "6:00 PM - 6:15 PM (Europe/Madrid)", attachment.Text)
}
