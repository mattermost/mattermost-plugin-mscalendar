// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {COMMAND} from '../support/constants';
import {eventSubject, loadCredentials} from '../support/credentials';
import {deleteEventsWithPrefix, waitForEvent} from '../support/graph';
import {ensureConnected, runSlashCommand} from '../support/ui';

const INVITE_TIMEOUT_MS = 90_000;

function tomorrow() {
    const date = new Date(Date.now() + (24 * 60 * 60 * 1000));
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

test.describe('events', () => {
    const user1 = () => loadCredentials().user1;
    const user2 = () => loadCredentials().user2;

    test.afterAll(async () => {
        await Promise.all([user1(), user2()].map((user) => deleteEventsWithPrefix(user.email, eventSubject('create'))));
    });

    test('creates an event from the modal and invites the second user', async ({pw}) => {
        const subject = eventSubject('create');
        const channelsPage = await ensureConnected(pw);
        const page = channelsPage.page;

        await runSlashCommand(channelsPage, `${COMMAND} event create`);
        const form = page.getByTestId('mscalendar-create-event-form');
        await expect(form).toBeVisible();

        await form.locator('#subject').fill(subject);
        await form.locator('#date').fill(tomorrow());

        await form.locator('#start_time').click();
        await page.keyboard.type('10:00');
        await page.keyboard.press('Enter');

        // The attendee selector accepts plain email addresses.
        await form.locator('#attendees').click();
        await page.keyboard.type(user2().email);
        await page.keyboard.press('Enter');

        await form.locator('#submit-button').click();
        await expect(form).toBeHidden();
        await expect(page.getByTestId('mscalendar-create-event-error')).toHaveCount(0);

        const created = await waitForEvent(user1().email, subject);
        expect(created.subject).toBe(subject);

        // The invitation reaches the attendee's calendar through Exchange, which can lag.
        await waitForEvent(user2().email, subject, INVITE_TIMEOUT_MS);
    });

    test('keeps the create button disabled until the form is valid', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);
        const page = channelsPage.page;

        await runSlashCommand(channelsPage, `${COMMAND} event create`);
        const form = page.getByTestId('mscalendar-create-event-form');
        await expect(form).toBeVisible();
        await expect(form.locator('#submit-button')).toBeDisabled();

        await form.locator('#subject').fill(eventSubject('invalid'));
        await expect(form.locator('#submit-button')).toBeDisabled();
    });
});
