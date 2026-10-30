// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, test} from '@mattermost/playwright-lib';

import {loginAdmin} from '../support/admin';
import {CalendarSidebar} from '../support/calendar_sidebar';
import {COMMAND} from '../support/constants';
import {USER1} from '../support/credentials';
import {fake} from '../support/fake';
import {disconnectRegularUser, patchPluginSettings} from '../support/plugin';
import {commandError, ensureConnected, ensureDisconnected, openConnectPopup, runCommandExpecting, submitSlashCommand} from '../support/ui';

test.describe('sign-in failures', () => {
    /**
     * @objective Verify that a wrong Microsoft password shows an error and leaves the account disconnected
     */
    test('rejects a wrong password and stays disconnected', {tag: '@mscalendar_sign_in'}, async ({pw}) => {
        // # Open the connect popup and submit a wrong password
        const channelsPage = await ensureDisconnected(pw);
        const login = await openConnectPopup(channelsPage);
        await login.submitCredentials({email: USER1.email, password: 'not-the-password'});

        // * Verify Microsoft shows a password error
        await expect(login.error).toBeVisible();

        // # Close the popup
        await login.page.close();

        // * Verify the sidebar still offers to connect
        await expect(new CalendarSidebar(channelsPage).connectButton).toBeVisible();
    });

    test.describe('consent', () => {
        test.afterEach(async () => {
            const {client} = await loginAdmin();
            await patchPluginSettings(client, {oauth2forceconsent: false});
        });

        /**
         * @objective Verify that denying the consent screen leaves the account disconnected
         *
         * @precondition
         * The plugin setting that forces the consent screen is enabled.
         */
        test('stays disconnected when the user denies consent', {tag: '@mscalendar_sign_in'}, async ({pw}) => {
            // # Force the consent screen
            const {client} = await loginAdmin();
            await patchPluginSettings(client, {oauth2forceconsent: true});

            // # Sign in and cancel the consent screen
            const channelsPage = await ensureDisconnected(pw);
            const login = await openConnectPopup(channelsPage);
            await login.submitCredentials(USER1);
            await login.cancelConsentButton.click();
            await login.page.waitForLoadState();
            await login.page.close();

            // * Verify the sidebar still offers to connect
            await expect(new CalendarSidebar(channelsPage).connectButton).toBeVisible();
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

    /**
     * @objective Verify that the plugin refreshes an expiring access token without user action
     *
     * @precondition
     * The fake issues access tokens that fall inside the plugin's refresh window shortly after sign-in.
     */
    test('refreshes an expired access token transparently', {tag: '@mscalendar_tokens'}, async ({pw}) => {
        // # Connect and wait until the access token is due for refresh
        const channelsPage = await ensureConnected(pw);
        await channelsPage.page.waitForTimeout(PAST_EXPIRY_MS);
        await fake.clearRequests();

        // * Verify a command still works
        await runCommandExpecting(channelsPage, `${COMMAND} viewcal`, /Times are shown in|no upcoming events/);

        // * Verify the plugin used the refresh token grant
        const grants = (await fake.requests()).map((request) => request.grant);
        expect(grants).toContain('refresh_token');
    });

    /**
     * @objective Verify that commands fail cleanly when Microsoft rejects the refresh token
     *
     * @precondition
     * The fake issues access tokens that fall inside the plugin's refresh window shortly after sign-in.
     */
    test('fails the command when the refresh token is revoked', {tag: '@mscalendar_tokens'}, async ({pw}) => {
        // # Connect, revoke the refresh token and wait until a refresh is due
        const channelsPage = await ensureConnected(pw);
        await fake.revokeRefreshTokens();
        await channelsPage.page.waitForTimeout(PAST_EXPIRY_MS);
        await fake.clearRequests();

        // # Run a command that needs a fresh token
        await submitSlashCommand(channelsPage, `${COMMAND} viewcal`);

        // * Verify the command reports an error
        await expect(commandError(channelsPage, 'Unable to execute command')).toBeVisible();
        await channelsPage.centerView.postCreate.input.fill('');

        // * Verify every refresh attempt was rejected
        const refreshes = (await fake.requests()).filter((request) => request.grant === 'refresh_token');
        expect(refreshes.length).toBeGreaterThan(0);
        expect(refreshes.every((request) => request.status === 400)).toBe(true);
    });
});
