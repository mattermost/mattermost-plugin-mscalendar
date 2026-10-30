// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test, type ChannelsPage} from '@mattermost/playwright-lib';

import {BotSettings, botMessage, openBotDirectMessage, openTownSquareChannel} from '../support/bot_dm';
import {eventSubject, USER1, USER2} from '../support/credentials';
import {deleteEventsWithPrefix, fake, waitForEvent} from '../support/fake';
import {ensureConnected} from '../support/ui';

const HOUR_MS = 60 * 60 * 1000;
const NOTIFICATIONS_SETTING = 'Receive notifications of new events';

async function subscription() {
    return (await fake.subscriptions())[0];
}

async function setNotifications(channelsPage: ChannelsPage, value: 'Yes' | 'No') {
    const settings = new BotSettings(channelsPage);
    await settings.open();
    await settings.chooseButton(NOTIFICATIONS_SETTING, value);
    await openTownSquareChannel(channelsPage);
}

test.describe('webhook notifications', () => {
    test.beforeEach(async ({pw}) => {
        const channelsPage = await ensureConnected(pw);
        await setNotifications(channelsPage, 'Yes');
        await expect.poll(async () => (await fake.subscriptions()).length).toBe(1);
    });

    test.afterEach(async ({pw}) => {
        const channelsPage = await ensureConnected(pw);
        await setNotifications(channelsPage, 'No');
        await deleteEventsWithPrefix(USER1.email, eventSubject('webhook'));
        await deleteEventsWithPrefix(USER2.email, eventSubject('webhook'));
    });

    /**
     * @objective Verify that enabling notifications creates a Graph subscription secured with a client state
     */
    test('registers a subscription that the fake validated and secured with a client state', {tag: '@mscalendar_webhooks'}, async () => {
        // * Verify the subscription has a client state and points at the plugin's webhook
        const sub = await subscription();
        expect(sub.clientState).not.toBe('');
        expect(sub.notificationUrl).toContain('/notification/v1/event');
    });

    /**
     * @objective Verify that the bot sends a direct message for a new invitation and for its update
     */
    test('sends a DM when an invitation arrives and when it is updated', {tag: '@mscalendar_webhooks'}, async ({pw}) => {
        const subject = eventSubject('webhook invite');
        const channelsPage = await ensureConnected(pw);
        const start = new Date(Date.now() + (4 * HOUR_MS));

        // # Invite the connected user to an event
        await fake.createEvent(USER2.email, subject, start, new Date(start.getTime() + HOUR_MS), [USER1.email]);
        const invited = await waitForEvent(USER1.email, subject);

        // * Verify the bot announces the new event
        await openBotDirectMessage(channelsPage);
        await expect(botMessage(channelsPage, `(new) ${subject}`)).toBeVisible({timeout: 20_000});

        // # Rename the event
        await fake.updateEvent(USER1.email, invited.id, {subject: `${subject} moved`});

        // * Verify the bot announces the update
        await expect(botMessage(channelsPage, `(updated) ${subject} moved`)).toBeVisible({timeout: 20_000});
    });

    /**
     * @objective Verify that the plugin ignores a change notification whose client state does not match
     */
    test('ignores a notification with the wrong client state', {tag: '@mscalendar_webhooks'}, async ({pw}) => {
        const subject = eventSubject('webhook forged');
        const channelsPage = await ensureConnected(pw);
        const start = new Date(Date.now() + (5 * HOUR_MS));

        // # Create an event and wait for the genuine notification to make the plugin fetch it
        const created = await fake.createEvent(USER1.email, subject, start, new Date(start.getTime() + HOUR_MS), [USER2.email]);
        const eventFetches = async () => (await fake.requests()).filter((request) => request.method === 'GET' && request.path?.toLowerCase().endsWith(`/events/${created.id.toLowerCase()}`));
        await expect.poll(async () => (await eventFetches()).length).toBeGreaterThan(0);
        await fake.clearRequests();

        // # Send a notification with a forged client state
        await fake.notify((await subscription()).id, {eventId: created.id, clientState: 'forged'});
        await channelsPage.page.waitForTimeout(3000);

        // * Verify the plugin did not fetch the event again
        expect(await eventFetches()).toHaveLength(0);
    });

    /**
     * @objective Verify that the plugin renews its subscription when a notification recommends it
     */
    test('renews the subscription when Microsoft recommends it', {tag: '@mscalendar_webhooks'}, async () => {
        const sub = await subscription();
        const start = new Date(Date.now() + (6 * HOUR_MS));
        const event = await fake.createEvent(USER1.email, eventSubject('webhook renew'), start, new Date(start.getTime() + HOUR_MS), [USER2.email]);
        const renewals = async () => (await fake.requests()).filter((request) => request.method === 'PATCH' && request.status === 200 && request.path?.endsWith(`/subscriptions/${sub.id}`));

        // # Send a notification that says the subscription expires soon
        await fake.notify(sub.id, {eventId: event.id, soon: true});

        // * Verify the plugin renewed the subscription
        await expect.poll(async () => (await renewals()).length).toBeGreaterThan(0);
    });
});
