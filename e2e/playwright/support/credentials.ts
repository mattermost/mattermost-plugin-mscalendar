// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './env';

export type TestUser = {email: string; password: string};

// These match the defaults baked into fake-microsoft/server.ts.
export const FAKE_TENANT_ID = 'fake-tenant';
export const FAKE_CLIENT_ID = 'fake-client-id';
export const FAKE_CLIENT_SECRET = 'fake-client-secret';
export const USER1: TestUser = {email: 'alice@contoso.test', password: 'Alice-Fake-Pass1'};
export const USER2: TestUser = {email: 'bob@contoso.test', password: 'Bob-Fake-Pass1'};

export function currentRunId(): string {
    return process.env.MSCAL_E2E_RUN_ID || '';
}

export function startRun(): string {
    const runId = currentRunId() || Date.now().toString(36);
    process.env.MSCAL_E2E_RUN_ID = runId;
    return runId;
}

export function eventSubject(title: string): string {
    return `[e2e-${currentRunId()}] ${title}`;
}
