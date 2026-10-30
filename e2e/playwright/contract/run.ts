// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// Runs the same Graph scenarios against the fake and/or the live tenant and checks that
// each response satisfies the contract and that both sides agree on status codes.
//   npm run contract:fake      fake only (needs the compose stack, or node fake-microsoft/server.ts)
//   npm run contract:live      live tenant only (needs MSCAL_E2E_* secrets)
//   npm run contract:compare   both, and fail when they disagree
import '../support/env';

import {missingFields} from '../fake-microsoft/contract.ts';

type Target = {name: string; graph: string; tokenUrl: string; clientId: string; clientSecret: string; user1: string; user2: string};
type Step = {name: string; status: number; violations: string[]};

const FAKE_URL = process.env.FAKE_URL || 'http://127.0.0.1:8080';

function fakeTarget(): Target {
    return {
        name: 'fake',
        graph: `${FAKE_URL}/v1.0`,
        tokenUrl: `${FAKE_URL}/fake-tenant/oauth2/v2.0/token`,
        clientId: 'fake-client-id',
        clientSecret: 'fake-client-secret',
        user1: 'alice@contoso.test',
        user2: 'bob@contoso.test',
    };
}

function liveTarget(): Target {
    const required = (name: string) => {
        const value = process.env[name];
        if (!value) {
            throw new Error(`Set ${name} to run the live contract checks (see e2e/playwright/README.md).`);
        }
        return value;
    };
    return {
        name: 'live',
        graph: 'https://graph.microsoft.com/v1.0',
        tokenUrl: `https://login.microsoftonline.com/${encodeURIComponent(required('MSCAL_E2E_TENANT_ID'))}/oauth2/v2.0/token`,
        clientId: required('MSCAL_E2E_CLIENT_ID'),
        clientSecret: required('MSCAL_E2E_CLIENT_SECRET'),
        user1: required('MSCAL_E2E_USER1_EMAIL'),
        user2: required('MSCAL_E2E_USER2_EMAIL'),
    };
}

async function appToken(target: Target): Promise<{status: number; body: any}> {
    const response = await fetch(target.tokenUrl, {
        method: 'POST',
        body: new URLSearchParams({client_id: target.clientId, client_secret: target.clientSecret, scope: 'https://graph.microsoft.com/.default', grant_type: 'client_credentials'}),
    });
    return {status: response.status, body: await response.json()};
}

async function call(target: Target, token: string, method: string, path: string, body?: unknown) {
    const response = await fetch(path.startsWith('http') ? path : `${target.graph}${path}`, {
        method,
        headers: {Authorization: `Bearer ${token}`, 'Content-Type': 'application/json'},
        body: body === undefined ? undefined : JSON.stringify(body),
    });
    const text = await response.text();
    return {status: response.status, body: text ? JSON.parse(text) : undefined};
}

function check(name: string, status: number, contract: string | undefined, body: unknown): Step {
    const violations = contract ? missingFields(contract, body).map((f) => `missing ${f}`) : [];
    return {name, status, violations};
}

async function runScenarios(target: Target): Promise<Step[]> {
    const steps: Step[] = [];
    const runId = `${Date.now().toString(36)}`;
    const subject = `[contract-${runId}] contract check`;
    let eventId = '';

    const tokenResult = await appToken(target);
    steps.push(check('client credentials token', tokenResult.status, 'appToken', tokenResult.body));
    const token = tokenResult.body.access_token as string;

    try {
        const start = new Date(Date.now() + (3 * 60 * 60 * 1000));
        const end = new Date(start.getTime() + (30 * 60 * 1000));
        const payload = {
            subject,
            start: {dateTime: start.toISOString().replace('Z', ''), timeZone: 'UTC'},
            end: {dateTime: end.toISOString().replace('Z', ''), timeZone: 'UTC'},
            attendees: [{emailAddress: {address: target.user2}, type: 'required'}],
        };
        const created = await call(target, token, 'POST', `/users/${encodeURIComponent(target.user1)}/events`, payload);
        steps.push(check('create event', created.status, 'event', created.body));
        eventId = created.body?.id || '';

        if (eventId) {
            const fetched = await call(target, token, 'GET', `/users/${encodeURIComponent(target.user1)}/events/${encodeURIComponent(eventId)}`);
            steps.push(check('get event', fetched.status, 'event', fetched.body));

            const viewQuery = `startDateTime=${encodeURIComponent(start.toISOString())}&endDateTime=${encodeURIComponent(end.toISOString())}&$top=20`;
            const view = await call(target, token, 'GET', `/users/${encodeURIComponent(target.user1)}/calendarView?${viewQuery}`);
            const found = (view.body?.value || []).find((e: any) => e.id === eventId);
            steps.push(check('calendar view contains the event', view.status, 'event', found));

            const batch = await call(target, token, 'POST', '/$batch', {
                requests: [{id: target.user1, method: 'GET', url: `/Users/${encodeURIComponent(target.user1)}/calendarView?${viewQuery}`, headers: {}}],
            });
            steps.push(check('batch calendar view envelope', batch.status, undefined, batch.body));
            const inner = batch.body?.responses?.[0];
            const batchFound = (inner?.body?.value || []).find((e: any) => e.id === eventId);
            steps.push(check('batch calendar view item', inner?.status ?? 0, 'event', batchFound));
        } else {
            const skipped = {status: 0, violations: ['skipped, no event was created']};
            for (const name of ['get event', 'calendar view contains the event', 'batch calendar view envelope', 'batch calendar view item']) {
                steps.push({name, ...skipped});
            }
        }

        const settings = await call(target, token, 'GET', `/users/${encodeURIComponent(target.user1)}/mailboxSettings`);
        steps.push(check('mailbox settings', settings.status, 'mailboxSettings', settings.body));

        const calendars = await call(target, token, 'GET', `/users/${encodeURIComponent(target.user1)}/calendars`);
        steps.push(check('calendars', calendars.status, 'calendar', calendars.body?.value?.[0]));

        const me = await call(target, token, 'GET', '/me');
        steps.push(check('/me with an app-only token is rejected', me.status, 'graphError', me.body));

        const bad = await call(target, 'not-a-token', 'GET', `/users/${encodeURIComponent(target.user1)}/events`);
        steps.push(check('invalid token is rejected', bad.status, 'graphError', bad.body));
    } finally {
        if (eventId) {
            const removed = await call(target, token, 'DELETE', `/users/${encodeURIComponent(target.user1)}/events/${encodeURIComponent(eventId)}`);
            steps.push(check('delete event', removed.status, undefined, undefined));
            const gone = await call(target, token, 'GET', `/users/${encodeURIComponent(target.user1)}/events/${encodeURIComponent(eventId)}`);
            steps.push(check('deleted event is not found', gone.status, 'graphError', gone.body));
        }
    }

    return steps;
}

function report(target: Target, steps: Step[]): boolean {
    let ok = true;
    for (const step of steps) {
        const bad = step.violations.length > 0;
        ok = ok && !bad;
        // eslint-disable-next-line no-console
        console.log(`${bad ? 'FAIL' : 'ok  '} [${target.name}] ${step.name} -> ${step.status}${bad ? ` (${step.violations.join(', ')})` : ''}`);
    }
    return ok;
}

async function main() {
    const mode = process.argv[2] || 'fake';
    const targets = mode === 'compare' ? [fakeTarget(), liveTarget()] : [mode === 'live' ? liveTarget() : fakeTarget()];

    const results: Array<{target: Target; steps: Step[]}> = [];
    let ok = true;
    for (const target of targets) {
        // eslint-disable-next-line no-await-in-loop
        const steps = await runScenarios(target);
        results.push({target, steps});
        ok = report(target, steps) && ok;
    }

    if (results.length === 2) {
        const [fake, live] = results;
        fake.steps.forEach((step, i) => {
            const other = live.steps[i];
            if (!other || other.name !== step.name || other.status !== step.status) {
                ok = false;
                // eslint-disable-next-line no-console
                console.log(`DRIFT ${step.name}: fake answered ${step.status}, live answered ${other?.status}`);
            }
        });
    }

    process.exit(ok ? 0 : 1);
}

main().catch((err) => {
    // eslint-disable-next-line no-console
    console.error(err);
    process.exit(1);
});
