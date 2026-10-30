// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {expect, type ChannelsPage} from '@mattermost/playwright-lib';
import type {Locator} from '@playwright/test';

import {BOT_USERNAME, COMMAND, TEAM_NAME, TOWN_SQUARE} from './constants';
import {runSlashCommand} from './ui';

export async function openBotDirectMessage(channelsPage: ChannelsPage) {
    await channelsPage.page.goto(`/${TEAM_NAME}/messages/@${BOT_USERNAME}`);
    await channelsPage.toBeVisible();
    await expect(channelsPage.centerView.postCreate.input).toBeVisible();
}

export async function openTownSquareChannel(channelsPage: ChannelsPage) {
    await channelsPage.goto(TEAM_NAME, TOWN_SQUARE);
    await channelsPage.toBeVisible();
}

export function botMessage(channelsPage: ChannelsPage, text: string) {
    return channelsPage.centerView.container.getByText(text).first();
}

// The settings command posts every setting as a block in one interactive bot message.
export class BotSettings {
    readonly channelsPage: ChannelsPage;

    constructor(channelsPage: ChannelsPage) {
        this.channelsPage = channelsPage;
    }

    async open() {
        await runSlashCommand(this.channelsPage, `${COMMAND} settings`);
        await openBotDirectMessage(this.channelsPage);
    }

    // Blocks have no role, so pick the innermost element holding both the title and its controls.
    setting(title: string): Locator {
        const page = this.channelsPage.page;
        const controls = page.getByRole('button').or(page.getByRole('combobox'));
        return this.channelsPage.centerView.container.
            locator('div').
            filter({has: page.getByText(`Setting: ${title}`, {exact: true})}).
            filter({has: controls}).
            last();
    }

    async chooseButton(title: string, value: string) {
        const setting = this.setting(title);
        await setting.getByRole('button', {name: value, exact: true}).click();
        await expect(setting).toContainText(`Current value: ${value}`);
    }

    async chooseOption(title: string, option: string) {
        const setting = this.setting(title);
        await setting.getByRole('combobox').click();
        await this.channelsPage.page.getByRole('option', {name: option, exact: true}).click();
        await expect(setting).toContainText(option);
    }
}
