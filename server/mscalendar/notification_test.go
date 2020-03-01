// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar/mock_plugin_api"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/remote/mock_remote"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/store/mock_store"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot/mock_bot"
)

func NewTestNotificationProcessor(env Env) NotificationProcessor {
	processor := &notificationProcessor{
		Env: env,
	}
	return processor
}

func TestProcessNotification(t *testing.T) {

	type webhook struct {
		ChangeType                     string `json:"changeType"`
		ClientState                    string `json:"clientState,omitempty"`
		Resource                       string `json:"resource,omitempty"`
		SubscriptionExpirationDateTime string `json:"subscriptionExpirationDateTime,omitempty"`
		SubscriptionID                 string `json:"subscriptionId"`
		ResourceData                   struct {
			DataType string `json:"@odata.type"`
		} `json:"resourceData"`
	}

	token := &oauth2.Token{
		AccessToken: "creator_oauth_token",
	}

	eRemote := &remote.Event{
		ID: "remote_event_id",
		Organizer: &remote.Attendee{
			EmailAddress: &remote.EmailAddress{
				Address: "event_organizer_email",
				Name:    "event_organizer_name",
			},
		},
		Location: &remote.Location{
			DisplayName: "event_location_display_name",
		},
		ResponseStatus: &remote.EventResponseStatus{
			Response: "event_response",
		},
		Weblink:           "event_weblink",
		Subject:           "event_subject",
		BodyPreview:       "event_body_preview",
		ResponseRequested: true,
	}

	sStore := &store.Subscription{
		PluginVersion: "x.x.x",
		Remote: &remote.Subscription{
			ID:          "remote_subscription_id",
			ClientState: "stored_client_state",
		},
		MattermostCreatorID: "creator_mm_id",
	}

	uStore := &store.User{
		Settings: store.Settings{
			EventSubscriptionID: "remote_subscription_id",
		},
		Remote:           &remote.User{},
		OAuth2Token:      token,
		MattermostUserID: "creator_mm_id",
	}

	nRemote := &remote.Notification{
		SubscriptionID: "remote_subscription_id",
		IsBare:         true,
		Webhook: &webhook{
			Resource:       "remote_event_resource_location",
			SubscriptionID: "sub_id",
			ChangeType:     "created",
			ResourceData: struct {
				DataType string `json:"@odata.type"`
			}{
				DataType: "#Microsoft.Graph.Event",
			},
		},
		SubscriptionCreator: &remote.User{},
		Event:               &remote.Event{},
		Subscription:        &remote.Subscription{},
	}

	tcs := []struct {
		name                string
		incomingClientState string
		expectedError       string
		hasPriorEvent       bool
		recommendRenew      bool
	}{
		{
			name:                "incoming ClientState matches stored ClientState",
			incomingClientState: "stored_client_state",
			expectedError:       "",
			hasPriorEvent:       false,
			recommendRenew:      false,
		}, {
			name:                "incoming ClientState doesn't match stored ClientState",
			incomingClientState: "wrong_client_state",
			expectedError:       "Unauthorized webhook",
			hasPriorEvent:       false,
			recommendRenew:      false,
		}, {
			name:                "prior event exists",
			incomingClientState: "stored_client_state",
			expectedError:       "",
			hasPriorEvent:       true,
			recommendRenew:      false,
		}, {
			name:                "sub renewal recommended",
			incomingClientState: "stored_client_state",
			expectedError:       "",
			hasPriorEvent:       false,
			recommendRenew:      true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock_store.NewMockStore(ctrl)
			logger := &bot.TestLogger{TB: t}
			mockPoster := mock_bot.NewMockPoster(ctrl)
			mockRemote := mock_remote.NewMockRemote(ctrl)
			mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)
			mockClient := mock_remote.NewMockClient(ctrl)

			conf := &config.Config{BotUserID: "bot_mm_id", PluginVersion: "x.x.x"}
			env := Env{
				Config: conf,
				Dependencies: &Dependencies{
					Store:     mockStore,
					Logger:    logger,
					Poster:    mockPoster,
					Remote:    mockRemote,
					PluginAPI: mockPluginAPI,
				},
			}

			nRemote.ClientState = tc.incomingClientState
			nRemote.RecommendRenew = tc.recommendRenew

			nRemoteWithEvent := *nRemote
			nRemoteWithEvent.Event = eRemote

			mockStore.EXPECT().LoadSubscription("remote_subscription_id").Return(sStore, nil).Times(1)
			mockStore.EXPECT().LoadUser("creator_mm_id").Return(uStore, nil).Times(1)

			if nRemote.ClientState == sStore.Remote.ClientState {
				mockRemote.EXPECT().MakeClient(context.Background(), token).Return(mockClient).Times(1)

				if tc.recommendRenew {
					mockClient.EXPECT().RenewSubscription("remote_subscription_id").Return(&remote.Subscription{}, nil).Times(1)
					mockStore.EXPECT().StoreUserSubscription(uStore, &store.Subscription{
						Remote:              &remote.Subscription{},
						MattermostCreatorID: uStore.MattermostUserID,
						PluginVersion:       conf.PluginVersion,
					}).Return(nil).Times(1)
				}

				mockClient.EXPECT().GetNotificationData(nRemote).Return(&nRemoteWithEvent, nil).Times(1)

				if tc.hasPriorEvent {
					priorEvent := *eRemote
					priorEvent.Location = &remote.Location{
						DisplayName:  "prior_event_location_display_name",
					}
					mockStore.EXPECT().LoadUserEvent("creator_mm_id", "remote_event_id").Return(&store.Event{
						Remote: &priorEvent,
					}, nil).Times(1)
				} else {
					mockStore.EXPECT().LoadUserEvent("creator_mm_id", "remote_event_id").Return(nil, store.ErrNotFound).Times(1)
				}

				mockPoster.EXPECT().DMWithAttachments("creator_mm_id", gomock.Any()).Return(nil).Times(1)
				mockStore.EXPECT().StoreUserEvent("creator_mm_id", gomock.Any()).Return(nil).Times(1)
			}

			p := NewTestNotificationProcessor(env)
			processor := p.(*notificationProcessor)
			err := processor.processNotification(nRemote)

			if tc.expectedError != "" {
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}
