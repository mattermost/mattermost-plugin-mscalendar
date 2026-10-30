// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {loginRegularUser} from '../support/admin';
import {BotSettings, openTownSquareChannel} from '../support/bot_dm';
import {COMMAND} from '../support/constants';
import {eventSubject, loadCredentials} from '../support/credentials';
import {createEvent, deleteEventsWithPrefix} from '../support/fake';
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

    /**
     * @objective Verify that the availability sync sets Do Not Disturb while a meeting is running
     *
     * @precondition
     * The sync reads the user's real status, so no other run may use the same mailbox at the same time.
     */
    test('sets Do Not Disturb from a running event', {tag: '@mscalendar_availability'}, async ({pw}) => {
        // # Connect the account and choose Do Not Disturb as the in-meeting status
        const channelsPage = await ensureConnected(pw);
        const settings = new BotSettings(channelsPage);
        await settings.open();
        await settings.chooseOption('Update Status', 'Do Not Disturb');

        // # Create a running event; the plugin ignores events without attendees
        await createEvent(user1().email, subject(), new Date(Date.now() - (5 * MINUTE_MS)), new Date(Date.now() + (25 * MINUTE_MS)), [loadCredentials().user2.email]);

        // # Reset the status and run the availability sync
        const {client, user} = await loginRegularUser();
        await client.updateStatus({user_id: user.id, status: 'online'});
        await openTownSquareChannel(channelsPage);
        await runSlashCommand(channelsPage, `${COMMAND} avail`);

        // * Verify the status changed to Do Not Disturb
        await expect.poll(async () => (await client.getStatus(user.id)).status, {timeout: 30_000}).toBe('dnd');
    });
});
