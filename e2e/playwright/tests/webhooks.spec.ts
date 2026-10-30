// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {BOT_USERNAME, COMMAND, TEAM_NAME} from '../support/constants';
import {eventSubject, USER1, USER2} from '../support/credentials';
import {deleteEventsWithPrefix, fake, waitForEvent} from '../support/fake';
import {ensureConnected, runSlashCommand} from '../support/ui';

const HOUR_MS = 60 * 60 * 1000;
const NOTIFICATIONS_SETTING = 'Setting: Receive notifications of new events';

async function subscription() {
    return (await fake.subscriptions())[0];
}

async function setNotifications(channelsPage: Awaited<ReturnType<typeof ensureConnected>>, value: 'Yes' | 'No') {
    const page = channelsPage.page;
    await runSlashCommand(channelsPage, `${COMMAND} settings`);
    await page.goto(`/${TEAM_NAME}/messages/@${BOT_USERNAME}`);
    const setting = page.getByText(NOTIFICATIONS_SETTING).last().locator('xpath=ancestor::div[.//button][1]');
    await setting.getByRole('button', {name: value, exact: true}).click();
    await expect(setting).toContainText(`Current value: ${value}`);
    await page.goto(`/${TEAM_NAME}/channels/town-square`);
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

    test('registers a subscription that the fake validated and secured with a client state', async () => {
        const sub = await subscription();
        expect(sub.clientState).not.toBe('');
        expect(sub.notificationUrl).toContain('/notification/v1/event');
    });

    test('sends a DM when an invitation arrives and when it is updated', async ({pw}) => {
        const subject = eventSubject('webhook invite');
        const channelsPage = await ensureConnected(pw);
        const start = new Date(Date.now() + (4 * HOUR_MS));

        await fake.createEvent(USER2.email, subject, start, new Date(start.getTime() + HOUR_MS), [USER1.email]);
        const invited = await waitForEvent(USER1.email, subject);

        await channelsPage.page.goto(`/${TEAM_NAME}/messages/@${BOT_USERNAME}`);
        await expect(channelsPage.page.getByText(`(new) ${subject}`).first()).toBeVisible({timeout: 20_000});

        await fake.updateEvent(USER1.email, invited.id, {subject: `${subject} moved`});
        await expect(channelsPage.page.getByText(`(updated) ${subject} moved`).first()).toBeVisible({timeout: 20_000});
    });

    test('ignores a notification with the wrong client state', async ({pw}) => {
        const subject = eventSubject('webhook forged');
        const channelsPage = await ensureConnected(pw);
        const start = new Date(Date.now() + (5 * HOUR_MS));

        const created = await fake.createEvent(USER1.email, subject, start, new Date(start.getTime() + HOUR_MS), [USER2.email]);
        await fake.clearRequests();
        await fake.notify((await subscription()).id, {eventId: created.id, clientState: 'forged'});

        await channelsPage.page.waitForTimeout(3000);
        const fetched = (await fake.requests()).filter((request) => request.method === 'GET' && request.path?.includes('/events/'));
        expect(fetched).toHaveLength(0);
    });

    test('renews the subscription when Microsoft recommends it', async () => {
        const sub = await subscription();
        const before = sub.expirationDateTime;
        const start = new Date(Date.now() + (6 * HOUR_MS));
        const event = await fake.createEvent(USER1.email, eventSubject('webhook renew'), start, new Date(start.getTime() + HOUR_MS), [USER2.email]);

        await fake.notify(sub.id, {eventId: event.id, soon: true});

        await expect.poll(async () => (await subscription()).expirationDateTime).not.toBe(before);
    });
});
