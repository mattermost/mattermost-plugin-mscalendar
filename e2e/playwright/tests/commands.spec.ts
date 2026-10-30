// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {test} from '@mattermost/playwright-lib';

import {COMMAND} from '../support/constants';
import {currentRunId, eventSubject, loadCredentials} from '../support/credentials';
import {createEvent, deleteEventsWithPrefix} from '../support/fake';
import {ensureConnected, runCommandExpecting} from '../support/ui';

const MINUTE_MS = 60 * 1000;
const AGENDA_REPLY = /Times are shown in|no upcoming events|no events for that day/;

test.describe('commands', () => {
    const user1 = () => loadCredentials().user1;

    test.beforeAll(async () => {
        const start = new Date(Date.now() + (10 * MINUTE_MS));
        await createEvent(user1().email, eventSubject('commands fixture'), start, new Date(start.getTime() + (30 * MINUTE_MS)));
    });

    test.afterAll(async () => {
        await deleteEventsWithPrefix(user1().email, eventSubject('commands fixture'));
    });

    /**
     * @objective Verify that the help and info subcommands reply with their content
     */
    test('prints help and plugin info', {tag: '@mscalendar_commands'}, async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        // # Run help and info
        // * Verify each reply
        await runCommandExpecting(channelsPage, `${COMMAND} help`, 'Read help text for the commands');
        await runCommandExpecting(channelsPage, `${COMMAND} info`, /plugin version/i);
    });

    /**
     * @objective Verify that connect reports the linked Microsoft account when already connected
     */
    test('reports that the account is already connected', {tag: '@mscalendar_commands'}, async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        // # Run connect
        // * Verify the reply names the connected mailbox
        await runCommandExpecting(channelsPage, `${COMMAND} connect`, user1().email);
    });

    /**
     * @objective Verify that the today and tomorrow subcommands reply with an agenda
     */
    test('shows the agenda for today and tomorrow', {tag: '@mscalendar_commands'}, async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        // # Run today and tomorrow
        // * Verify each reply is an agenda
        await runCommandExpecting(channelsPage, `${COMMAND} today`, AGENDA_REPLY);
        await runCommandExpecting(channelsPage, `${COMMAND} tomorrow`, AGENDA_REPLY);
    });

    /**
     * @objective Verify that viewcal lists an upcoming Microsoft event
     */
    test('lists upcoming events with viewcal', {tag: '@mscalendar_commands'}, async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        // # Run viewcal
        // * Verify the reply contains this run's fixture event
        await runCommandExpecting(channelsPage, `${COMMAND} viewcal`, `e2e-${currentRunId()}`);
    });

    /**
     * @objective Verify that summary settings replies with the daily summary configuration
     */
    test('shows the daily summary settings', {tag: '@mscalendar_commands'}, async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        // # Run summary settings
        // * Verify the reply describes the daily summary
        await runCommandExpecting(channelsPage, `${COMMAND} summary settings`, /daily summary/i);
    });

    /**
     * @objective Verify that showcals lists the connected user's calendars
     */
    test('lists the calendars as a plugin admin', {tag: '@mscalendar_commands'}, async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        // # Run showcals
        // * Verify the reply contains calendar data
        await runCommandExpecting(channelsPage, `${COMMAND} showcals`, /"name"/);
    });

    /**
     * @objective Verify that an unknown subcommand falls back to the help text
     */
    test('falls back to help for unknown subcommands', {tag: '@mscalendar_commands'}, async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        // # Run an unknown subcommand
        // * Verify the reply is the help text
        await runCommandExpecting(channelsPage, `${COMMAND} nonsense`, 'Read help text for the commands');
    });
});
