// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './env';

import fs from 'node:fs';
import path from 'node:path';

export type TestUser = {email: string; password: string};

// These match the defaults baked into fake-microsoft/server.ts.
export const FAKE_TENANT_ID = 'fake-tenant';
export const FAKE_CLIENT_ID = 'fake-client-id';
export const FAKE_CLIENT_SECRET = 'fake-client-secret';
export const USER1: TestUser = {email: 'alice@contoso.test', password: 'Alice-Fake-Pass1'};
export const USER2: TestUser = {email: 'bob@contoso.test', password: 'Bob-Fake-Pass1'};

const RUN_STATE_FILE = path.resolve(__dirname, '../results/run-state.json');

export function currentRunId(): string {
    return process.env.MSCAL_E2E_RUN_ID || '';
}

export function startRun(): string {
    const runId = currentRunId() || Date.now().toString(36);
    process.env.MSCAL_E2E_RUN_ID = runId;
    fs.mkdirSync(path.dirname(RUN_STATE_FILE), {recursive: true});
    fs.writeFileSync(RUN_STATE_FILE, JSON.stringify({runId}));
    return runId;
}

export function lastRunId(): string | undefined {
    if (currentRunId()) {
        return currentRunId();
    }
    try {
        return JSON.parse(fs.readFileSync(RUN_STATE_FILE, 'utf8')).runId;
    } catch {
        return undefined;
    }
}

export function eventSubject(title: string): string {
    return `[e2e-${currentRunId()}] ${title}`;
}

export function loadCredentials() {
    return {user1: USER1, user2: USER2};
}
