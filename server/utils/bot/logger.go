// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package bot

import "fmt"

import "github.com/mattermost/mattermost-plugin-msoffice/server/utils"

type LogContext map[string]interface{}

func level(l string) int {
	switch l {
	case "debug":
		return 4
	case "info":
		return 3
	case "warn":
		return 2
	case "error":
		return 1
	}
	return 0
}

type Logger interface {
	With(LogContext) Logger
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

func toKeyValuePairs(in map[string]interface{}) (out []interface{}) {
	for k, v := range in {
		out = append(out, k)
		out = append(out, v)
	}
	return out
}

func (bot *bot) With(logContext LogContext) Logger {
	newbot := *bot
	if len(newbot.logContext) == 0 {
		newbot.logContext = map[string]interface{}{}
	}
	for k, v := range logContext {
		newbot.logContext[k] = v
	}
	return &newbot
}

func (bot *bot) Debugf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	bot.pluginAPI.LogDebug(message, toKeyValuePairs(bot.logContext)...)
	if level(bot.AdminLogLevel) >= 4 {
		bot.logToAdmins("DEBUG", message)
	}
}

func (bot *bot) Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	bot.pluginAPI.LogError(message, toKeyValuePairs(bot.logContext)...)
	if level(bot.AdminLogLevel) >= 1 {
		bot.logToAdmins("ERROR", message)
	}
}

func (bot *bot) Infof(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	bot.pluginAPI.LogInfo(message, toKeyValuePairs(bot.logContext)...)
	if level(bot.AdminLogLevel) >= 3 {
		bot.logToAdmins("INFO", message)
	}
}

func (bot *bot) Warnf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	bot.pluginAPI.LogWarn(message, toKeyValuePairs(bot.logContext)...)
	if level(bot.AdminLogLevel) >= 2 {
		bot.logToAdmins("WARN", message)
	}
}

func (bot *bot) logToAdmins(level, message string) {
	if bot.AdminLogVerbose && len(bot.logContext) > 0 {
		message += "\n" + utils.JSONBlock(bot.logContext)
	}
	bot.dmAdmins("(log " + level + ") " + message)
}

type NilLogger struct{}

func (l *NilLogger) With(logContext LogContext) Logger         { return l }
func (l *NilLogger) Debugf(format string, args ...interface{}) {}
func (l *NilLogger) Errorf(format string, args ...interface{}) {}
func (l *NilLogger) Infof(format string, args ...interface{})  {}
func (l *NilLogger) Warnf(format string, args ...interface{})  {}
