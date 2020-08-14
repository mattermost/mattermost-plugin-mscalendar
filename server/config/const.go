// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

const (
	BotUserName    = "mscalendar"
	BotDisplayName = "Microsoft Calendar"
	BotDescription = "Created by the Microsoft Calendar Plugin."

	ApplicationName    = "Microsoft Calendar"
	Repository         = "mattermost-plugin-mscalendar"
	CommandTrigger     = "mscalendar"
	TelemetryShortName = "mscalendar"

	PathOAuth2              = "/oauth2"
	PathComplete            = "/complete"
	PathAPI                 = "/api/v1"
	PathPostAction          = "/action"
	PathRespond             = "/respond"
	PathAccept              = "/accept"
	PathDecline             = "/decline"
	PathTentative           = "/tentative"
	PathConfirmStatusChange = "/confirm"
	PathNotification        = "/notification/v1"
	PathEvent               = "/event"

	FullPathEventNotification = PathNotification + PathEvent
	FullPathOAuth2Redirect    = PathOAuth2 + PathComplete

	EventIDKey = "EventID"
)
