// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {CalendarSidebar} from '../support/calendar_sidebar';
import {eventSubject, loadCredentials} from '../support/credentials';
import {createEvent, deleteEventsWithPrefix} from '../support/fake';
import {ensureConnected, openTownSquare} from '../support/ui';

const MINUTE_MS = 60 * 1000;

test.describe('experimental calendar sidebar', () => {
    const user1 = () => loadCredentials().user1;
    const subject = () => eventSubject('sidebar fixture');

    test.afterAll(async () => {
        await deleteEventsWithPrefix(user1().email, subject());
    });

    /**
     * @objective Verify that the plugin registers its app bar button and that it opens the calendar sidebar
     */
    test('registers the channel header button and the sidebar', {tag: '@mscalendar_sidebar'}, async ({pw}) => {
        const channelsPage = await openTownSquare(pw);
        const sidebar = new CalendarSidebar(channelsPage);

        // * Verify the app bar button is visible
        await expect(sidebar.appBarIcon).toBeVisible();

        // # Open the sidebar
        // * Verify it shows either the calendar or the connect prompt
        await sidebar.open();
    });

    /**
     * @objective Verify that an event from the Microsoft calendar appears in the sidebar
     */
    test('shows a Microsoft event in the sidebar', {tag: '@mscalendar_sidebar'}, async ({pw}) => {
        // # Create an upcoming event and open the connected sidebar
        const start = new Date(Date.now() + (10 * MINUTE_MS));
        await createEvent(user1().email, subject(), start, new Date(start.getTime() + (30 * MINUTE_MS)));
        const channelsPage = await ensureConnected(pw);

        // * Verify the event is shown
        await expect(new CalendarSidebar(channelsPage).event(subject())).toBeVisible({timeout: 30_000});
    });
});
