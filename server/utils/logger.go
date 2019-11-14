// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package utils

type Logger interface {
	LogDebug(message string, keyValuePairs ...interface{})
	LogError(message string, keyValuePairs ...interface{})
	LogInfo(message string, keyValuePairs ...interface{})
	LogWarn(message string, keyValuePairs ...interface{})
}

var NilLogger Logger = &nilLogger{}

type nilLogger struct{}

func (l *nilLogger) LogDebug(message string, keyValuePairs ...interface{}) {}
func (l *nilLogger) LogError(message string, keyValuePairs ...interface{}) {}
func (l *nilLogger) LogInfo(message string, keyValuePairs ...interface{})  {}
func (l *nilLogger) LogWarn(message string, keyValuePairs ...interface{})  {}
