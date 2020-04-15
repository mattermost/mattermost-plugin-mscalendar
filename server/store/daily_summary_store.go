// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/kvstore"
)

type DailySummaryStore interface {
	LoadDailySummaryIndex() (DailySummaryIndex, error)
	LoadDailySummaryUserSettings(mattermostUserID string) (*DailySummaryUserSettings, error)
	StoreDailySummaryUserSettings(dsum *DailySummaryUserSettings) error
	DeleteDailySummaryUserSettings(mattermostUserID string) error
	ModifyDailySummaryIndex(modify func(dsumIndex DailySummaryIndex) (DailySummaryIndex, error)) error
}

type DailySummaryUserSettings struct {
	MattermostUserID string `json:"mm_id"`
	RemoteID         string `json:"remote_id"`
	Enable           bool   `json:"enable"`
	PostTime         string `json:"post_time"` // Kitchen format, i.e. 8:30AM
	Timezone         string `json:"tz"`        // Timezone in MSCal when PostTime is set/updated
	LastPostTime     string `json:"last_post_time"`
}

type DailySummaryIndex []*DailySummaryUserSettings

func (s *pluginStore) LoadDailySummaryIndex() (DailySummaryIndex, error) {
	dsumIndex := DailySummaryIndex{}
	err := kvstore.LoadJSON(s.dailySummaryKV, "", &dsumIndex)
	if err != nil && err.Error() != "not found" {
		return nil, err
	}
	return dsumIndex, nil
}

func (s *pluginStore) LoadDailySummaryUserSettings(mattermostUserID string) (*DailySummaryUserSettings, error) {
	index, err := s.LoadDailySummaryIndex()
	if err != nil {
		return nil, err
	}
	if index == nil {
		return nil, nil
	}

	for _, dsum := range index {
		if dsum.MattermostUserID == mattermostUserID {
			return dsum, nil
		}
	}
	return nil, nil
}

func (s *pluginStore) StoreDailySummaryUserSettings(toStore *DailySummaryUserSettings) error {
	return s.ModifyDailySummaryIndex(func(dsumIndex DailySummaryIndex) (DailySummaryIndex, error) {
		for i, dsum := range dsumIndex {
			if dsum.MattermostUserID == toStore.MattermostUserID {
				result := append(dsumIndex[:i], toStore)
				return append(result, dsumIndex[i+1:]...), nil
			}
		}

		return append(dsumIndex, toStore), nil
	})
}

func (s *pluginStore) DeleteDailySummaryUserSettings(mattermostUserID string) error {
	return s.ModifyDailySummaryIndex(func(dsumIndex DailySummaryIndex) (DailySummaryIndex, error) {
		for i, u := range dsumIndex {
			if u.MattermostUserID == mattermostUserID {
				return append(dsumIndex[:i], dsumIndex[i+1:]...), nil
			}
		}
		return dsumIndex, nil
	})
}

func (s *pluginStore) ModifyDailySummaryIndex(modify func(dsumIndex DailySummaryIndex) (DailySummaryIndex, error)) error {
	return kvstore.AtomicModify(s.dailySummaryKV, "", func(initialBytes []byte, storeErr error) ([]byte, error) {
		if storeErr != nil && storeErr != ErrNotFound {
			return nil, storeErr
		}

		storedSettings := DailySummaryIndex{}
		if len(initialBytes) > 0 {
			err := json.Unmarshal(initialBytes, &storedSettings)
			if err != nil {
				return nil, err
			}
		}

		updated, err := modify(storedSettings)
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(updated)
		if err != nil {
			return nil, err
		}

		return b, nil
	})
}

func (index DailySummaryIndex) ByRemoteID() map[string]*DailySummaryUserSettings {
	result := map[string]*DailySummaryUserSettings{}
	for _, u := range index {
		result[u.RemoteID] = u
	}
	return result
}
