// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

// REVIEW: need an interface for returning bot info
// probably good to have a struct to capture the data clump
const (
	BotDescription = "Created by the %s Plugin."

	PathOAuth2                = "/oauth2"
	PathComplete              = "/complete"
	PathAPI                   = "/api/v1"
	PathDialogs               = "/dialogs"
	PathSetAutoRespondMessage = "/set-auto-respond-message"
	PathPostAction            = "/action"
	PathRespond               = "/respond"
	PathAccept                = "/accept"
	PathDecline               = "/decline"
	PathTentative             = "/tentative"
	PathConfirmStatusChange   = "/confirm"
	PathNotification          = "/notification/v1"
	PathEvent                 = "/event"
	PathVerifyDomain          = "/verify"

	PathAutocomplete = "/autocomplete"
	PathUsers        = "/users"
	PathChannels     = "/channels"

	InternalAPIPath   = "/api/v1"
	PathEvents        = "/events"
	PathCreate        = "/create"
	PathProvider      = "/provider"
	PathConnectedUser = "/me"

	FullPathEventNotification = PathNotification + PathEvent
	FullPathOAuth2Redirect    = PathOAuth2 + PathComplete

	EventIDKey = "EventID"
)
