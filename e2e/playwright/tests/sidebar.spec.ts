// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {APP_BAR_ICON} from '../support/constants';
import {eventSubject, loadCredentials} from '../support/credentials';
import {createEvent, deleteEventsWithPrefix} from '../support/graph';
import {ensureConnected, openSidebar, openTownSquare} from '../support/ui';

const MINUTE_MS = 60 * 1000;

test.describe('experimental calendar sidebar', () => {
    const user1 = () => loadCredentials().user1;
    const subject = () => eventSubject('sidebar fixture');

    test.afterAll(async () => {
        await deleteEventsWithPrefix(user1().email, subject());
    });

    test('registers the channel header button and the sidebar', async ({pw}) => {
        const channelsPage = await openTownSquare(pw);

        await expect(channelsPage.page.locator(APP_BAR_ICON)).toBeVisible();
        await openSidebar(channelsPage.page);
    });

    test('shows a Microsoft event in the sidebar', async ({pw}) => {
        const start = new Date(Date.now() + (10 * MINUTE_MS));
        await createEvent(user1().email, subject(), start, new Date(start.getTime() + (30 * MINUTE_MS)));

        const channelsPage = await ensureConnected(pw);
        const sidebar = channelsPage.page.getByTestId('mscalendar-sidebar');

        await expect(sidebar.locator('.fc-event').filter({hasText: subject()})).toBeVisible({timeout: 30_000});
    });
});
