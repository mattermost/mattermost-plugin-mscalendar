// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package mscalendar

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

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

func newTestNotificationProcessor(env Env) NotificationProcessor {
	processor := &notificationProcessor{
		Env: env,
	}
	return processor
}

func newTestEvent(locationDisplayName string, subjectDisplayName string) *remote.Event {
	return &remote.Event{
		ID:      "remote_event_id",
		ICalUID: "remote_event_uid",
		Organizer: &remote.Attendee{
			EmailAddress: &remote.EmailAddress{
				Address: "event_organizer_email",
				Name:    "event_organizer_name",
			},
		},
		Location: &remote.Location{
			DisplayName: locationDisplayName,
		},
		ResponseStatus: &remote.EventResponseStatus{
			Response: "event_response",
		},
		Weblink:           "event_weblink",
		Subject:           subjectDisplayName,
		BodyPreview:       "event_body_preview",
		ResponseRequested: true,
	}
}

func newTestSubscription() *store.Subscription {
	return &store.Subscription{
		PluginVersion: "x.x.x",
		Remote: &remote.Subscription{
			ID:          "remote_subscription_id",
			ClientState: "stored_client_state",
			CreatorID:   "remote_user_id",
		},
		MattermostCreatorID: "creator_mm_id",
	}
}

func newTestUser() *store.User {
	return &store.User{
		Settings: store.Settings{
			EventSubscriptionID: "remote_subscription_id",
		},
		Remote: &remote.User{ID: "remote_user_id"},
		OAuth2Token: &oauth2.Token{
			AccessToken: "creator_oauth_token",
		},
		MattermostUserID: "creator_mm_id",
	}
}

func newTestNotification(clientState string, recommendRenew bool) *remote.Notification {
	n := &remote.Notification{
		SubscriptionID:      "remote_subscription_id",
		IsBare:              true,
		SubscriptionCreator: &remote.User{},
		Event:               newTestEvent("event_location_display_name", "event_subject"),
		Subscription:        &remote.Subscription{},
		ClientState:         clientState,
		RecommendRenew:      recommendRenew,
	}
	return n
}

func TestProcessNotification(t *testing.T) {
	tcs := []struct {
		notification  *remote.Notification
		priorEvent    *remote.Event
		name          string
		expectedError string
	}{
		{
			name:          "incoming ClientState matches stored ClientState",
			expectedError: "",
			notification:  newTestNotification("stored_client_state", false),
			priorEvent:    nil,
		}, {
			name:          "incoming ClientState doesn't match stored ClientState",
			expectedError: "unauthorized webhook",
			notification:  newTestNotification("wrong_client_state", false),
			priorEvent:    nil,
		}, {
			name:          "prior event exists",
			expectedError: "",
			notification:  newTestNotification("stored_client_state", false),
			priorEvent:    newTestEvent("prior_event_location_display_name", "other_event_subject"),
		}, {
			name:          "sub renewal recommended",
			expectedError: "",
			notification:  newTestNotification("stored_client_state", true),
			priorEvent:    nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock_store.NewMockStore(ctrl)
			mockPoster := mock_bot.NewMockPoster(ctrl)
			mockRemote := mock_remote.NewMockRemote(ctrl)
			mockPluginAPI := mock_plugin_api.NewMockPluginAPI(ctrl)
			mockClient := mock_remote.NewMockClient(ctrl)

			conf := &config.Config{PluginVersion: "x.x.x"}
			env := Env{
				Config: conf,
				Dependencies: &Dependencies{
					Store:     mockStore,
					Logger:    &bot.NilLogger{},
					Poster:    mockPoster,
					Remote:    mockRemote,
					PluginAPI: mockPluginAPI,
				},
			}

			subscription := newTestSubscription()
			user := newTestUser()

			mockStore.EXPECT().LoadSubscription("remote_subscription_id").Return(subscription, nil).Times(1)
			mockStore.EXPECT().LoadUser("creator_mm_id").Return(user, nil).Times(1)

			if tc.notification.ClientState == subscription.Remote.ClientState {
				mockRemote.EXPECT().MakeClient(context.Background(), &oauth2.Token{
					AccessToken: "creator_oauth_token",
				}).Return(mockClient).Times(1)
				mockClient.EXPECT().GetMailboxSettings(user.Remote.ID).Return(&remote.MailboxSettings{TimeZone: "Eastern Standard Time"}, nil)

				if tc.notification.RecommendRenew {
					mockClient.EXPECT().RenewSubscription("remote_subscription_id").Return(&remote.Subscription{}, nil).Times(1)
					mockStore.EXPECT().StoreUserSubscription(user, &store.Subscription{
						Remote:              &remote.Subscription{},
						MattermostCreatorID: "creator_mm_id",
						PluginVersion:       "x.x.x",
					}).Return(nil).Times(1)
				}

				mockClient.EXPECT().GetNotificationData(tc.notification).Return(tc.notification, nil).Times(1)

				if tc.priorEvent != nil {
					mockStore.EXPECT().LoadUserEvent("creator_mm_id", "remote_event_uid").Return(&store.Event{
						Remote: tc.priorEvent,
					}, nil).Times(1)
				} else {
					mockStore.EXPECT().LoadUserEvent("creator_mm_id", "remote_event_uid").Return(nil, store.ErrNotFound).Times(1)
				}

				mockPoster.EXPECT().DMWithAttachments("creator_mm_id", gomock.Any()).Return("", nil).Times(1)
				mockStore.EXPECT().StoreUserEvent("creator_mm_id", gomock.Any()).Return(nil).Times(1)
			}

			p := newTestNotificationProcessor(env)
			processor := p.(*notificationProcessor)
			err := processor.processNotification(tc.notification)

			if tc.expectedError != "" {
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestProcessNotificationOverflow(t *testing.T) {
	t.Run("overflow", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		processor := &notificationProcessor{
			queue: make(chan (*remote.Notification), maxQueueSize),
		}

		for i := 0; i < maxQueueSize; i++ {
			err := processor.Enqueue(&remote.Notification{})
			require.NoError(t, err)
		}
		err := processor.Enqueue(&remote.Notification{})
		require.Error(t, err)
	})
}
