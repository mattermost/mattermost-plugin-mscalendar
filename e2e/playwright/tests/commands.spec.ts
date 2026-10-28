// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {test} from '@mattermost/playwright-lib';

import {COMMAND} from '../support/constants';
import {currentRunId, eventSubject, loadCredentials} from '../support/credentials';
import {createEvent, deleteEventsWithPrefix} from '../support/graph';
import {ensureConnected, runCommandExpecting} from '../support/ui';

const MINUTE_MS = 60 * 1000;

test.describe('commands', () => {
    const user1 = () => loadCredentials().user1;

    test.beforeAll(async () => {
        const start = new Date(Date.now() + (10 * MINUTE_MS));
        await createEvent(user1().email, eventSubject('commands fixture'), start, new Date(start.getTime() + (30 * MINUTE_MS)));
    });

    test.afterAll(async () => {
        await deleteEventsWithPrefix(user1().email, eventSubject('commands fixture'));
    });

    test('prints help and plugin info', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        await runCommandExpecting(channelsPage, `${COMMAND} help`, 'Read help text for the commands');
        await runCommandExpecting(channelsPage, `${COMMAND} info`, /plugin version/i);
    });

    test('reports that the account is already connected', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        await runCommandExpecting(channelsPage, `${COMMAND} connect`, user1().email);
    });

    test('shows the agenda for today and tomorrow', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        await runCommandExpecting(channelsPage, `${COMMAND} today`, /Agenda for|no events for that day/);
        await runCommandExpecting(channelsPage, `${COMMAND} tomorrow`, /Agenda for|no events for that day/);
    });

    test('lists upcoming events with viewcal', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        await runCommandExpecting(channelsPage, `${COMMAND} viewcal`, `e2e-${currentRunId()}`);
    });

    test('shows the daily summary settings', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        await runCommandExpecting(channelsPage, `${COMMAND} summary settings`, /daily summary/i);
    });

    test('lists the calendars as a plugin admin', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        await runCommandExpecting(channelsPage, `${COMMAND} showcals`, /"name"/);
    });

    test('falls back to help for unknown subcommands', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);

        await runCommandExpecting(channelsPage, `${COMMAND} nonsense`, 'Read help text for the commands');
    });
});
