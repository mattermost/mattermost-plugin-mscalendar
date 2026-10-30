// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {loginAdmin} from '../support/admin';
import {COMMAND} from '../support/constants';
import {USER1} from '../support/credentials';
import {fake} from '../support/fake';
import {submitCredentials} from '../support/ms_login';
import {disconnectRegularUser, patchPluginSettings} from '../support/plugin';
import {ensureConnected, ensureDisconnected, openConnectPopup, runCommandExpecting} from '../support/ui';

test.describe('sign-in failures', () => {
    test('rejects a wrong password and stays disconnected', async ({pw}) => {
        const channelsPage = await ensureDisconnected(pw);
        const popup = await openConnectPopup(channelsPage.page);

        await submitCredentials(popup, {email: USER1.email, password: 'not-the-password'});
        await expect(popup.locator('#passwordError')).toBeVisible();

        await popup.close();
        await expect(channelsPage.page.getByTestId('mscalendar-sidebar-connect')).toBeVisible();
    });

    test.describe('consent', () => {
        test.afterEach(async () => {
            const {client} = await loginAdmin();
            await patchPluginSettings(client, {oauth2forceconsent: false});
        });

        test('stays disconnected when the user denies consent', async ({pw}) => {
            const {client} = await loginAdmin();
            await patchPluginSettings(client, {oauth2forceconsent: true});

            const channelsPage = await ensureDisconnected(pw);
            const popup = await openConnectPopup(channelsPage.page);
            await submitCredentials(popup, USER1);

            await popup.locator('#cancelConsent').click();
            await popup.waitForLoadState();
            await popup.close();

            await expect(channelsPage.page.getByTestId('mscalendar-sidebar-connect')).toBeVisible();
        });
    });
});

test.describe('token lifecycle', () => {
    // The plugin refreshes tokens that expire within 5 minutes, so a 330s token needs a short wait.
    const SHORT_TTL_SECONDS = 330;
    const PAST_EXPIRY_MS = 35_000;

    test.beforeEach(async ({pw}) => {
        await ensureDisconnected(pw);
        await fake.configure({accessTokenTtlSeconds: SHORT_TTL_SECONDS});
    });

    test.afterEach(async () => {
        await fake.configure({accessTokenTtlSeconds: 3600});
        await disconnectRegularUser();
    });

    test('refreshes an expired access token transparently', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);
        await channelsPage.page.waitForTimeout(PAST_EXPIRY_MS);
        await fake.clearRequests();

        await runCommandExpecting(channelsPage, `${COMMAND} viewcal`, /Times are shown in|no upcoming events/);

        const grants = (await fake.requests()).map((request) => request.grant);
        expect(grants).toContain('refresh_token');
    });

    test('fails the command when the refresh token is revoked', async ({pw}) => {
        const channelsPage = await ensureConnected(pw);
        await fake.revokeRefreshTokens();
        await channelsPage.page.waitForTimeout(PAST_EXPIRY_MS);

        await fake.clearRequests();

        const input = channelsPage.centerView.postCreate.input;
        await input.fill(`${COMMAND} viewcal`);
        await channelsPage.centerView.postCreate.sendMessageButton.click();
        await expect(channelsPage.page.getByText('Unable to execute command')).toBeVisible();
        await input.fill('');

        const refreshes = (await fake.requests()).filter((request) => request.grant === 'refresh_token');
        expect(refreshes.length).toBeGreaterThan(0);
        expect(refreshes.every((request) => request.status === 400)).toBe(true);
    });
});
