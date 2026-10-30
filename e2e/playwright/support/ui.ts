// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, type ChannelsPage, type PlaywrightExtended} from '@mattermost/playwright-lib';
import type {BrowserContext, Page} from '@playwright/test';

import {loginRegularUser} from './admin';
import {APP_BAR_ICON, COMMAND, TEAM_NAME, TOWN_SQUARE} from './constants';
import {USER1} from './credentials';
import {FAKE_URL} from './fake';
import {signInToMicrosoft} from './ms_login';

const SIDEBAR_CONNECT = 'mscalendar-sidebar-connect';
const SIDEBAR = 'mscalendar-sidebar';

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

// The app-bar icon is rendered by Mattermost, so it has no data-testid from this plugin.
export async function openSidebar(page: Page) {
    const connect = page.getByTestId(SIDEBAR_CONNECT);
    const sidebar = page.getByTestId(SIDEBAR);
    if (await connect.isVisible().catch(() => false) || await sidebar.isVisible().catch(() => false)) {
        return;
    }
    await page.locator(APP_BAR_ICON).click();
    await expect(connect.or(sidebar)).toBeVisible();
}

export async function runSlashCommand(channelsPage: ChannelsPage, command: string) {
    const postCreate = channelsPage.centerView.postCreate;
    await postCreate.input.fill(command);

    // Pressing Enter would pick a slash-command autocomplete suggestion.
    await postCreate.sendMessageButton.click();
    await expect(postCreate.input).toHaveValue('');
}

// Earlier commands can leave an identical reply on screen, so wait for one more match.
export async function runCommandExpecting(channelsPage: ChannelsPage, command: string, reply: string | RegExp) {
    const replies = channelsPage.page.locator('.post-message__text').filter({hasText: reply});
    const before = await replies.count();
    await runSlashCommand(channelsPage, command);
    await expect(replies).toHaveCount(before + 1);
    return replies.last();
}

export async function connectAccount(page: Page) {
    const popup = await openConnectPopup(page);
    await signInToMicrosoft(popup, USER1);
    await expect(page.getByTestId(SIDEBAR)).toBeVisible({timeout: 30000});
}

export async function disconnectAccount(channelsPage: ChannelsPage) {
    await runCommandExpecting(channelsPage, `${COMMAND} disconnect`, 'Successfully disconnected your account');
    await expect(channelsPage.page.getByTestId(SIDEBAR_CONNECT)).toBeVisible();
}

export async function ensureConnected(pw: PlaywrightExtended) {
    const channelsPage = await openTownSquare(pw);
    await openSidebar(channelsPage.page);
    if (await channelsPage.page.getByTestId(SIDEBAR_CONNECT).isVisible().catch(() => false)) {
        await connectAccount(channelsPage.page);
    }
    await expect(channelsPage.page.getByTestId(SIDEBAR)).toBeVisible();
    return channelsPage;
}

export async function ensureDisconnected(pw: PlaywrightExtended) {
    const channelsPage = await openTownSquare(pw);
    await openSidebar(channelsPage.page);
    if (await channelsPage.page.getByTestId(SIDEBAR).isVisible().catch(() => false)) {
        await disconnectAccount(channelsPage);
    }
    return channelsPage;
}

// Context routes attach to a new popup asynchronously, so its first redirect can reach the
// real login host. Reloading the authorize URL once the popup exists goes through the route.
export async function openConnectPopup(page: Page) {
    const popupPromise = page.waitForEvent('popup');
    await page.getByTestId(SIDEBAR_CONNECT).click();
    const popup = await popupPromise;

    await popup.waitForURL(/^https:\/\/login\.microsoftonline\.com\/.*\/authorize/, {waitUntil: 'commit'});
    await popup.goto(popup.url());
    return popup;
}
