// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {test} from '@mattermost/playwright-lib';

import {loginAdmin} from '../support/admin';
import {COMMAND} from '../support/constants';
import {loadCredentials} from '../support/credentials';
import {patchPluginSettings} from '../support/plugin';
import {connectAccount, disconnectAccount, ensureDisconnected, runCommandExpecting} from '../support/ui';

/**
 * @objective Verify that a regular member can connect and disconnect a Microsoft account through the login popup
 */
test('connects and disconnects a regular member through the Microsoft login popup', {tag: '@mscalendar_connect'}, async ({pw}) => {
    // # Sign in through the connect popup
    const channelsPage = await ensureDisconnected(pw);
    await connectAccount(channelsPage);

    // * Verify the plugin reports the connected mailbox
    await runCommandExpecting(channelsPage, `${COMMAND} connect`, loadCredentials().user1.email);

    // # Disconnect
    // * Verify the sidebar offers to connect again
    await disconnectAccount(channelsPage);
});

test.describe('forced consent', () => {
    test.afterEach(async () => {
        const {client} = await loginAdmin();
        await patchPluginSettings(client, {oauth2forceconsent: false});
    });

    /**
     * @objective Verify that connecting succeeds when the consent screen is always shown
     *
     * @precondition
     * The plugin setting that forces the consent screen is enabled.
     */
    test('connects when the consent screen is always shown', {tag: '@mscalendar_connect'}, async ({pw}) => {
        // # Force the consent screen
        const {client} = await loginAdmin();
        await patchPluginSettings(client, {oauth2forceconsent: true});

        // # Sign in and accept the consent screen
        // * Verify the calendar sidebar appears
        const channelsPage = await ensureDisconnected(pw);
        await connectAccount(channelsPage);
    });
});
