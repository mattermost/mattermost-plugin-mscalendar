// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {loginRegularUser} from '../support/admin';
import {BOT_USERNAME, COMMAND, TEAM_NAME} from '../support/constants';
import {eventSubject, loadCredentials} from '../support/credentials';
import {createEvent, deleteEventsWithPrefix} from '../support/graph';
import {ensureConnected, runSlashCommand} from '../support/ui';

const MINUTE_MS = 60 * 1000;

test.describe('availability', () => {
    const user1 = () => loadCredentials().user1;
    const subject = () => eventSubject('availability fixture');

    test.afterAll(async () => {
        await deleteEventsWithPrefix(user1().email, subject());
        const {client, user} = await loginRegularUser();
        await client.updateStatus({user_id: user.id, status: 'online'});
    });

    // The sync reads the user's real status, so it must not overlap with another run on the same mailbox.
    test('sets Do Not Disturb from a running event', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);
        const page = channelsPage.page;

        await runSlashCommand(channelsPage, `${COMMAND} settings`);
        await page.goto(`/${TEAM_NAME}/messages/@${BOT_USERNAME}`);
        const setting = page.locator('.attachment').filter({hasText: 'Setting: Update Status'}).last();
        await expect(setting).toBeVisible();
        await setting.locator('input').first().click();
        await page.getByText('Do Not Disturb', {exact: true}).last().click();
        await expect(setting).toContainText('Do Not Disturb');

        await createEvent(user1().email, subject(), new Date(Date.now() - (5 * MINUTE_MS)), new Date(Date.now() + (25 * MINUTE_MS)));

        const {client, user} = await loginRegularUser();
        await client.updateStatus({user_id: user.id, status: 'online'});

        await page.goto(`/${TEAM_NAME}/channels/town-square`);
        await channelsPage.toBeVisible();
        await runSlashCommand(channelsPage, `${COMMAND} avail`);

        await expect.poll(async () => (await client.getStatus(user.id)).status, {timeout: 30_000}).toBe('dnd');
    });
});
