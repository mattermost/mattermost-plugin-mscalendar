// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/kvstore"
)

type DailySummaryStore interface {
	LoadDailySummaryIndex() (DailySummaryIndex, error)
	SaveDailySummaryIndex(dsumIndex DailySummaryIndex) error
}

type DailySummarySettings struct {
	MattermostUserID string `json:"mm_id"`
	Enable           bool   `json:"enable"`
	PostTime         string `json:"post_time"` // Kitchen format, i.e. 8:30AM
	Timezone         string `json:"tz"`        // Timezone in MSCal when PostTime is set/updated
}

type DailySummaryIndex []*DailySummarySettings

func (s *pluginStore) LoadDailySummaryIndex() (DailySummaryIndex, error) {
	dsumIndex := DailySummaryIndex{}
	err := kvstore.LoadJSON(s.dailySummaryKV, "", &dsumIndex)
	if err != nil && err.Error() != "not found" {
		return nil, err
	}
	return dsumIndex, nil
}

func (s *pluginStore) SaveDailySummaryIndex(dsumIndex DailySummaryIndex) error {
	err := kvstore.StoreJSON(s.dailySummaryKV, "", &dsumIndex)
	if err != nil {
		return err
	}
	return nil
}
