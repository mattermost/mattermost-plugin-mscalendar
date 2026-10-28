// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {test} from '@mattermost/playwright-lib';

import {loginAdmin} from '../support/admin';
import {COMMAND} from '../support/constants';
import {loadCredentials} from '../support/credentials';
import {patchPluginSettings} from '../support/plugin';
import {connectAccount, disconnectAccount, openSidebar, openTownSquare, runCommandExpecting} from '../support/ui';

test('connects and disconnects a regular member through the Microsoft login popup', async ({pw}) => {
    const channelsPage = await openTownSquare(pw);
    await openSidebar(channelsPage.page);

    if (await channelsPage.page.getByTestId('mscalendar-sidebar').isVisible().catch(() => false)) {
        await disconnectAccount(channelsPage);
    }

    await connectAccount(channelsPage.page);

    const {user1} = loadCredentials();
    await runCommandExpecting(channelsPage, `${COMMAND} connect`, user1.email);

    await disconnectAccount(channelsPage);
});

test.describe('forced consent', () => {
    test.afterEach(async () => {
        const {client} = await loginAdmin();
        await patchPluginSettings(client, {oauth2forceconsent: false});
    });

    test('connects when the consent screen is always shown', async ({pw}) => {
        const {client} = await loginAdmin();
        await patchPluginSettings(client, {oauth2forceconsent: true});

        const channelsPage = await openTownSquare(pw);
        await openSidebar(channelsPage.page);
        if (await channelsPage.page.getByTestId('mscalendar-sidebar').isVisible().catch(() => false)) {
            await disconnectAccount(channelsPage);
        }

        // signInToMicrosoft accepts the consent screen that prompt=consent adds.
        await connectAccount(channelsPage.page);
    });
});
