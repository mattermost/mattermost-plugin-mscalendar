// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package utils

type Logger interface {
	LogDebug(message string, keyValuePairs ...interface{})
	LogError(message string, keyValuePairs ...interface{})
	LogInfo(message string, keyValuePairs ...interface{})
	LogWarn(message string, keyValuePairs ...interface{})
}
