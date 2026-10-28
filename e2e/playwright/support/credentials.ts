// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './env';

import fs from 'node:fs';
import path from 'node:path';

export type TestUser = {email: string; password: string};

export type Credentials = {
    tenantId: string;
    clientId: string;
    clientSecret: string;
    user1: TestUser;
    user2: TestUser;
};

const RUN_STATE_FILE = path.resolve(__dirname, '../results/run-state.json');

function required(name: string): string {
    const value = process.env[name];
    if (!value) {
        throw new Error(`Set ${name} (see e2e/playwright/README.md).`);
    }
    return value;
}

export function loadCredentials(): Credentials {
    return {
        tenantId: required('MSCAL_E2E_TENANT_ID'),
        clientId: required('MSCAL_E2E_CLIENT_ID'),
        clientSecret: required('MSCAL_E2E_CLIENT_SECRET'),
        user1: {email: required('MSCAL_E2E_USER1_EMAIL'), password: required('MSCAL_E2E_USER1_PASSWORD')},
        user2: {email: required('MSCAL_E2E_USER2_EMAIL'), password: required('MSCAL_E2E_USER2_PASSWORD')},
    };
}

// CI sets MSCAL_E2E_RUN_ID so its always-run cleanup step can find this run's events.
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
