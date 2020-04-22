// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils"
)

func (api *api) preprocessAction(w http.ResponseWriter, req *http.Request) (mscal mscalendar.MSCalendar, user *mscalendar.User, eventID string, option string, postID string) {
	mattermostUserID := req.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		utils.SlackAttachmentError(w, "Error: not authorized")
		return nil, nil, "", "", ""
	}

	request := model.PostActionIntegrationRequestFromJson(req.Body)
	if request == nil {
		utils.SlackAttachmentError(w, "Error: invalid request")
		return nil, nil, "", "", ""
	}

	eventID, ok := request.Context[config.EventIDKey].(string)
	if !ok {
		utils.SlackAttachmentError(w, "Error: missing event ID")
		return nil, nil, "", "", ""
	}
	option, _ = request.Context["selected_option"].(string)
	mscal = mscalendar.New(api.Env, mattermostUserID)

	return mscal, mscalendar.NewUser(mattermostUserID), eventID, option, request.PostId
}

func (api *api) postActionAccept(w http.ResponseWriter, req *http.Request) {
	mscalendar, user, eventID, _, _ := api.preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := mscalendar.AcceptEvent(user, eventID)
	if err != nil {
		api.Logger.Warnf(err.Error())
		utils.SlackAttachmentError(w, "Error: Failed to accept event: "+err.Error())
		return
	}
}

func (api *api) postActionDecline(w http.ResponseWriter, req *http.Request) {
	mscalendar, user, eventID, _, _ := api.preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := mscalendar.DeclineEvent(user, eventID)
	if err != nil {
		utils.SlackAttachmentError(w, "Error: Failed to decline event: "+err.Error())
		return
	}
}

func (api *api) postActionTentative(w http.ResponseWriter, req *http.Request) {
	mscalendar, user, eventID, _, _ := api.preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := mscalendar.TentativelyAcceptEvent(user, eventID)
	if err != nil {
		utils.SlackAttachmentError(w, "Error: Failed to tentatively accept event: "+err.Error())
		return
	}
}

func (api *api) postActionRespond(w http.ResponseWriter, req *http.Request) {
	calendar, user, eventID, option, postID := api.preprocessAction(w, req)
	if eventID == "" {
		return
	}
	err := calendar.RespondToEvent(user, eventID, option)
	if err != nil && !strings.HasPrefix(err.Error(), "202") && !strings.HasPrefix(err.Error(), "404") {
		utils.SlackAttachmentError(w, "Error: Failed to respond to event: "+err.Error())
		return
	}

	p, err := api.PluginAPI.GetPost(postID)
	if err != nil {
		utils.SlackAttachmentError(w, "Error: Failed to update the post: "+err.Error())
		return
	}

	sas := p.Attachments()
	if len(sas) == 0 {
		utils.SlackAttachmentError(w, "Error: Failed to update the post: "+err.Error())
		return
	}

	sa := sas[0]

	if err == nil || strings.HasPrefix(err.Error(), "202") {
		sa.Fields = append(sa.Fields, &model.SlackAttachmentField{
			Title: "Response",
			Value: fmt.Sprintf("You have %s this event", prettyOption(option)),
			Short: false,
		})
	}

	sa.Actions = []*model.PostAction{}
	postResponse := model.PostActionIntegrationResponse{}
	model.ParseSlackAttachment(p, []*model.SlackAttachment{sa})

	postResponse.Update = p

	if err != nil && strings.HasPrefix(err.Error(), "404") {
		postResponse.EphemeralText = "Event has changed since this message. Please change your status directly on MS Calendar."
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(postResponse.ToJson())
}

func prettyOption(option string) string {
	switch option {
	case mscalendar.OptionYes:
		return "accepted"
	case mscalendar.OptionNo:
		return "declined"
	case mscalendar.OptionMaybe:
		return "tentatively accepted"
	default:
		return ""
	}
}

func (api *api) postActionConfirmStatusChange(w http.ResponseWriter, req *http.Request) {
	mattermostUserID := req.Header.Get("Mattermost-User-ID")
	if mattermostUserID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	response := model.PostActionIntegrationResponse{}
	post := &model.Post{}

	request := model.PostActionIntegrationRequestFromJson(req.Body)
	if request == nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	value, ok := request.Context["value"].(bool)
	if !ok {
		http.Error(w, `No recognizable value for property "value"`, http.StatusBadRequest)
		return
	}

	returnText := "The status has not been changed."
	if value {
		changeTo, ok := request.Context["change_to"]
		if !ok {
			http.Error(w, "No state to change", http.StatusBadRequest)
			return
		}
		stringChangeTo := changeTo.(string)
		prettyChangeTo, ok := request.Context["pretty_change_to"]
		if !ok {
			prettyChangeTo = changeTo
		}
		stringPrettyChangeTo := prettyChangeTo.(string)

		api.PluginAPI.UpdateMattermostUserStatus(mattermostUserID, stringChangeTo)
		returnText = fmt.Sprintf("The status has been changed to %s.", stringPrettyChangeTo)
	}

	sa := &model.SlackAttachment{
		Title: "Status Change",
		Text:  returnText,
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{sa})

	response.Update = post
	w.Header().Set("Content-Type", "application/json")
	w.Write(response.ToJson())
}
