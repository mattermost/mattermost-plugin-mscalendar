// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

const (
	BotUserName    = "gcal"
	BotDisplayName = "Google Calendar"
	BotDescription = "Created by the Google Calendar Plugin."

	ApplicationName    = "Google Calendar"
	Repository         = "mattermost-plugin-gcal"
	CommandTrigger     = "gcal"
	TelemetryShortName = "gcal"

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

	FullPathEventNotification = PathNotification + PathEvent
	FullPathOAuth2Redirect    = PathOAuth2 + PathComplete

	EventIDKey = "EventID"
)
