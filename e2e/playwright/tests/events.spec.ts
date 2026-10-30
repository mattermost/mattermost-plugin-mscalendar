// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {CreateEventModal} from '../support/create_event_modal';
import {eventSubject, loadCredentials} from '../support/credentials';
import {deleteEventsWithPrefix, waitForEvent} from '../support/fake';
import {ensureConnected} from '../support/ui';

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

    /**
     * @objective Verify that the create event modal creates a Microsoft event and invites the guest
     */
    test('creates an event from the modal and invites the second user', {tag: '@mscalendar_events'}, async ({pw}) => {
        const subject = eventSubject('create');
        const channelsPage = await ensureConnected(pw);
        const modal = new CreateEventModal(channelsPage);

        // # Open the modal and fill in the event
        await modal.open(channelsPage);
        await modal.subjectInput.fill(subject);
        await modal.dateInput.fill(tomorrow());
        await modal.chooseStartTime('10:00');
        await modal.addGuest(user2().email);

        // # Create the event
        await modal.createButton.click();

        // * Verify the modal closes without an error
        await expect(modal.container).toBeHidden();
        await expect(modal.error).toHaveCount(0);

        // * Verify the event is on the organizer's calendar
        const created = await waitForEvent(user1().email, subject);
        expect(created.subject).toBe(subject);

        // * Verify the invitation reaches the guest's calendar, which can lag behind
        await waitForEvent(user2().email, subject, INVITE_TIMEOUT_MS);
    });

    /**
     * @objective Verify that the create button stays disabled until the required fields are filled in
     */
    test('keeps the create button disabled until the form is valid', {tag: '@mscalendar_events'}, async ({pw}) => {
        const channelsPage = await ensureConnected(pw);
        const modal = new CreateEventModal(channelsPage);

        // # Open the modal
        await modal.open(channelsPage);

        // * Verify the create button is disabled
        await expect(modal.createButton).toBeDisabled();

        // # Fill in only the subject
        await modal.subjectInput.fill(eventSubject('invalid'));

        // * Verify the create button is still disabled
        await expect(modal.createButton).toBeDisabled();
    });
});
