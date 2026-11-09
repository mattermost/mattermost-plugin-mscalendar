// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, type ChannelsPage} from '@mattermost/playwright-lib';
import type {Locator, Page} from '@playwright/test';

import {COMMAND} from './constants';
import {runSlashCommand} from './ui';

export class CreateEventModal {
    readonly page: Page;
    readonly container: Locator;
    readonly subjectInput: Locator;
    readonly dateInput: Locator;
    readonly startTimeInput: Locator;
    readonly guestsInput: Locator;
    readonly createButton: Locator;
    readonly error: Locator;

    constructor(channelsPage: ChannelsPage) {
        this.page = channelsPage.page;
        this.container = this.page.getByRole('dialog').filter({has: this.page.getByRole('heading', {name: 'Create Calendar Event'})});
        this.subjectInput = this.container.getByRole('textbox', {name: 'Subject'});
        this.dateInput = this.container.getByRole('textbox', {name: 'Date'});
        this.startTimeInput = this.container.getByRole('combobox', {name: 'Start Time'});
        this.guestsInput = this.container.getByRole('combobox', {name: 'Guests'});
        this.createButton = this.container.getByRole('button', {name: 'Create', exact: true});
        this.error = this.container.getByRole('alert');
    }

    async open(channelsPage: ChannelsPage) {
        await runSlashCommand(channelsPage, `${COMMAND} event create`);
        await expect(this.container).toBeVisible();
    }

    async chooseStartTime(time: string) {
        await this.startTimeInput.click();
        await this.page.keyboard.type(time);
        await this.page.keyboard.press('Enter');
    }

    // The guest selector accepts plain email addresses once the user lookup has answered.
    async addGuest(email: string) {
        await this.guestsInput.click();
        const lookup = this.page.waitForResponse((response) => response.url().includes('/autocomplete/users'));
        await this.page.keyboard.type(email);
        await lookup;
        await this.page.keyboard.press('Enter');
        await expect(this.guest(email)).toBeVisible();
    }

    guest(email: string) {
        return this.container.getByText(email).first();
    }
}
