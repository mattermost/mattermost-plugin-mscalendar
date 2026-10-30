// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, type ChannelsPage} from '@mattermost/playwright-lib';
import type {Locator} from '@playwright/test';

import {APP_BAR_ICON} from './constants';

export class CalendarSidebar {
    readonly appBarIcon: Locator;
    readonly container: Locator;
    readonly connectButton: Locator;

    constructor(channelsPage: ChannelsPage) {
        const page = channelsPage.page;

        // Mattermost renders the app-bar button without an accessible name.
        this.appBarIcon = channelsPage.appBar.container.locator(APP_BAR_ICON);

        // The calendar has no landmark role, so it keeps a test id.
        this.container = page.getByTestId('mscalendar-sidebar');
        this.connectButton = page.getByRole('link', {name: 'Connect account'});
    }

    async open() {
        if (await this.isOpen()) {
            return;
        }
        await this.appBarIcon.click();
        await expect(this.connectButton.or(this.container)).toBeVisible();
    }

    async isOpen() {
        return await this.connectButton.isVisible() || this.container.isVisible();
    }

    async isConnected() {
        return this.container.isVisible();
    }

    event(subject: string) {
        return this.container.getByText(subject);
    }
}
