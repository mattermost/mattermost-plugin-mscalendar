// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, type ChannelsPage, type PlaywrightExtended} from '@mattermost/playwright-lib';
import type {BrowserContext} from '@playwright/test';

import {loginRegularUser} from './admin';
import {CalendarSidebar} from './calendar_sidebar';
import {COMMAND, TEAM_NAME, TOWN_SQUARE} from './constants';
import {USER1} from './credentials';
import {FAKE_URL} from './fake';
import {MicrosoftLoginPage} from './ms_login';

// The plugin sends the browser to the real login host, so serve it from the fake instead.
export async function routeMicrosoftToFake(context: BrowserContext) {
    await context.route(/^https:\/\/login\.microsoftonline\.com\//, async (route) => {
        const url = route.request().url().replace('https://login.microsoftonline.com', FAKE_URL);
        const response = await route.fetch({url, maxRedirects: 0});
        await route.fulfill({response});
    });
}

// Specs act as a regular member. The plugin treats this member as a plugin admin for admin-only commands.
export async function openTownSquare(pw: PlaywrightExtended) {
    const {user} = await loginRegularUser();
    const {channelsPage} = await pw.testBrowser.login(user);
    await routeMicrosoftToFake(channelsPage.page.context());
    await channelsPage.goto(TEAM_NAME, TOWN_SQUARE);
    await channelsPage.toBeVisible();
    return channelsPage;
}

// Clicking send rather than pressing Enter, which would pick a slash-command autocomplete suggestion.
export async function submitSlashCommand(channelsPage: ChannelsPage, command: string) {
    const postCreate = channelsPage.centerView.postCreate;
    await postCreate.input.fill(command);
    await postCreate.sendMessageButton.click();
}

export async function runSlashCommand(channelsPage: ChannelsPage, command: string) {
    await submitSlashCommand(channelsPage, command);
    await expect(channelsPage.centerView.postCreate.input).toHaveValue('');
}

// Earlier commands can leave an identical reply on screen, so wait for a new last post.
export async function runCommandExpecting(channelsPage: ChannelsPage, command: string, reply: string | RegExp) {
    const centerView = channelsPage.centerView;
    const before = await centerView.getLastPostID();
    await runSlashCommand(channelsPage, command);
    await expect.poll(() => centerView.getLastPostID()).not.toBe(before);

    const post = await centerView.getLastPost();
    await expect(post.body).toContainText(reply);
    return post;
}

export function commandError(channelsPage: ChannelsPage, message: string) {
    return channelsPage.page.getByText(message);
}

// Playwright routes never see redirect targets, so letting the popup follow the plugin's redirect
// would reach the real login host. Open it blank, resolve the redirect here, and navigate directly.
export async function openConnectPopup(channelsPage: ChannelsPage) {
    const page = channelsPage.page;
    await page.evaluate(() => {
        const open = window.open.bind(window);
        window.open = (url, ...rest) => {
            (window as unknown as {connectUrl: string}).connectUrl = new URL(String(url), window.location.href).href;
            return open('about:blank', ...rest);
        };
    });

    const popupPromise = page.waitForEvent('popup');
    await new CalendarSidebar(channelsPage).connectButton.click();
    const popup = await popupPromise;

    const connectUrl = await page.evaluate(() => (window as unknown as {connectUrl: string}).connectUrl);
    const connect = await page.request.get(connectUrl, {maxRedirects: 0});
    const authorizeUrl = connect.headers().location;
    if (connect.status() !== 302 || !authorizeUrl) {
        throw new Error(`Expected a redirect from ${connectUrl}, got ${connect.status()}`);
    }

    await popup.goto(authorizeUrl);
    return new MicrosoftLoginPage(popup);
}

export async function connectAccount(channelsPage: ChannelsPage) {
    const login = await openConnectPopup(channelsPage);
    await login.signIn(USER1);
    await expect(new CalendarSidebar(channelsPage).container).toBeVisible({timeout: 30_000});
}

export async function disconnectAccount(channelsPage: ChannelsPage) {
    await runCommandExpecting(channelsPage, `${COMMAND} disconnect`, 'Successfully disconnected your account');
    await expect(new CalendarSidebar(channelsPage).connectButton).toBeVisible();
}

export async function ensureConnected(pw: PlaywrightExtended) {
    const channelsPage = await openTownSquare(pw);
    const sidebar = new CalendarSidebar(channelsPage);
    await sidebar.open();
    if (!await sidebar.isConnected()) {
        await connectAccount(channelsPage);
    }
    await expect(sidebar.container).toBeVisible();
    return channelsPage;
}

export async function ensureDisconnected(pw: PlaywrightExtended) {
    const channelsPage = await openTownSquare(pw);
    const sidebar = new CalendarSidebar(channelsPage);
    await sidebar.open();
    if (await sidebar.isConnected()) {
        await disconnectAccount(channelsPage);
    }
    return channelsPage;
}
