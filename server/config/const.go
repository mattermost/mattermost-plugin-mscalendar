// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

const (
	BotUserName    = "msoffice"
	BotDisplayName = "Microsoft Office TODO"
	BotDescription = "Created by the Microsoft Office Plugin. TODO"

	ApplicationName = "Microsoft Office"
	Repository      = "mattermost-plugin-msoffice"
	CommandTrigger  = "msoffice"

	PathOAuth2       = "/oauth2"
	PathComplete     = "/complete"
	PathAPI          = "/api/v1"
	PathPostAction   = "/action"
	PathRespond      = "/respond"
	PathAccept       = "/accept"
	PathDecline      = "/decline"
	PathTentative    = "/tentative"
	PathNotification = "/notification/v1"
	PathEvent        = "/event"

	FullPathEventNotification = PathNotification + PathEvent
	FullPathOAuth2Redirect    = PathOAuth2 + PathComplete

	EventIDKey = "EventID"
)
