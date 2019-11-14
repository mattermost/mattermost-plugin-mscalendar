// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

const (
	BotUserName    = "msoffice"
	BotDisplayName = "Microsoft Office TODO"
	BotDescription = "Created by the Microsoft Office Plugin. TODO"

	ApplicationName    = "Microsoft Office"
	Repository         = "mattermost-plugin-msoffice"
	CommandTrigger     = "msoffice"
	OAuth2Path         = "/oauth2"
	OAuth2CompletePath = "/complete"
	APIPath            = "/api/v1"
	WebhookPath        = "/webhook/v1"
	WebhookEventPath   = "/event"

	EventWebhookFullPath   = WebhookPath + WebhookEventPath
	OAuth2RedirectFullPath = OAuth2Path + OAuth2CompletePath
)
