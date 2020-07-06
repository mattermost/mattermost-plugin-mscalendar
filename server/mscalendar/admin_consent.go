// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"github.com/google/uuid"
)

type AdminConsentChecker interface {
	CreateNewAdminConsentToken(authedUserID string) (string, error)
	VerifyAdminConsentToken(state, authedUserID string) error
}

func (m *mscalendar) CreateNewAdminConsentToken(authedUserID string) (string, error) {
	toReturn := "adminconsent_" + uuid.New().String()
	toStore := toReturn + "_" + authedUserID

	err := m.Store.StoreOAuth2State(toStore)
	if err != nil {
		return "", err
	}

	return toReturn, nil
}

func (m *mscalendar) VerifyAdminConsentToken(state, authedUserID string) error {
	stored := state + "_" + authedUserID
	err := m.Store.VerifyOAuth2State(stored)
	if err != nil {
		return err
	}

	return nil
}
