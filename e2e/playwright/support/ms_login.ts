// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Page} from '@playwright/test';

import type {TestUser} from './credentials';

const FLOW_TIMEOUT_MS = 45_000;

// Walks the Microsoft sign-in popup: email, password, then the optional
// "Stay signed in?" and consent screens, until the plugin closes the popup.
export async function signInToMicrosoft(popup: Page, user: TestUser) {
    await submitCredentials(popup, user);
    await finishSignIn(popup);
}

// Every Microsoft page uses the same primary button id.
export async function submitCredentials(popup: Page, user: TestUser) {
    const primary = popup.locator('#idSIButton9');

    await popup.locator('input[name="loginfmt"]').fill(user.email);
    await primary.click();

    await popup.locator('input[name="passwd"]').fill(user.password);
    await primary.click();
}

export async function finishSignIn(popup: Page) {
    const deadline = Date.now() + FLOW_TIMEOUT_MS;
    while (!popup.isClosed() && Date.now() < deadline) {
        // eslint-disable-next-line no-await-in-loop
        await handleOptionalScreen(popup);
        // eslint-disable-next-line no-await-in-loop
        await popup.waitForTimeout(500).catch(() => undefined);
    }

    if (!popup.isClosed()) {
        throw new Error(`Microsoft sign-in did not finish. Last page: ${await describePage(popup)}`);
    }
}

async function handleOptionalScreen(popup: Page) {
    try {
        const decline = popup.locator('#idBtn_Back');
        if (await decline.isVisible()) {
            // "Stay signed in?" keeps sessions out of later specs.
            await decline.click();
            return;
        }

        const consent = popup.getByText(/Permissions requested|Let this app access/i);
        if (await consent.first().isVisible()) {
            await popup.locator('#idSIButton9').click();
            return;
        }

        const error = popup.locator('#usernameError, #passwordError, #service_exception_message');
        if (await error.first().isVisible()) {
            throw new Error(`Microsoft sign-in failed: ${await error.first().innerText()}`);
        }
    } catch (err) {
        // The popup closing mid-check is the success path.
        if (!popup.isClosed()) {
            throw err;
        }
    }
}

async function describePage(popup: Page) {
    const text = await popup.locator('body').innerText().catch(() => '');
    return `${popup.url()} ${text.replace(/\s+/g, ' ').slice(0, 300)}`;
}