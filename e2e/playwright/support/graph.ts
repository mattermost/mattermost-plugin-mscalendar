// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {loadCredentials} from './credentials';
import {EVENT_PREFIX} from './constants';

const GRAPH = 'https://graph.microsoft.com/v1.0';

export type GraphEvent = {
    id: string;
    subject: string;
    createdDateTime: string;
    attendees?: Array<{emailAddress: {address: string}}>;
};

let cachedToken: {value: string; expiresAt: number} | undefined;

// App-only token for the test app. The tenant's calendars are reachable with the
// Calendars.ReadWrite application permission.
async function accessToken(): Promise<string> {
    if (cachedToken && cachedToken.expiresAt > Date.now() + 60_000) {
        return cachedToken.value;
    }

    const {tenantId, clientId, clientSecret} = loadCredentials();
    const response = await fetch(`https://login.microsoftonline.com/${encodeURIComponent(tenantId)}/oauth2/v2.0/token`, {
        method: 'POST',
        body: new URLSearchParams({
            client_id: clientId,
            client_secret: clientSecret,
            scope: 'https://graph.microsoft.com/.default',
            grant_type: 'client_credentials',
        }),
    });
    if (!response.ok) {
        throw new Error(`Graph token request failed: ${response.status} ${await response.text()}`);
    }

    const body = await response.json() as {access_token: string; expires_in: number};
    cachedToken = {value: body.access_token, expiresAt: Date.now() + (body.expires_in * 1000)};
    return cachedToken.value;
}

async function graph<T>(method: string, url: string, body?: unknown): Promise<T | undefined> {
    const response = await fetch(url.startsWith('http') ? url : `${GRAPH}${url}`, {
        method,
        headers: {
            Authorization: `Bearer ${await accessToken()}`,
            'Content-Type': 'application/json',
        },
        body: body === undefined ? undefined : JSON.stringify(body),
    });
    if (!response.ok) {
        throw new Error(`Graph ${method} ${url} failed: ${response.status} ${await response.text()}`);
    }

    return response.status === 204 ? undefined : await response.json() as T;
}

function userPath(email: string) {
    return `/users/${encodeURIComponent(email)}`;
}

export async function createEvent(email: string, subject: string, start: Date, end: Date, attendeeEmails: string[] = []): Promise<GraphEvent> {
    const event = await graph<GraphEvent>('POST', `${userPath(email)}/events`, {
        subject,
        start: {dateTime: start.toISOString().replace('Z', ''), timeZone: 'UTC'},
        end: {dateTime: end.toISOString().replace('Z', ''), timeZone: 'UTC'},
        attendees: attendeeEmails.map((address) => ({emailAddress: {address}, type: 'required'})),
    });
    return event as GraphEvent;
}

export async function listEventsWithPrefix(email: string, prefix: string = EVENT_PREFIX): Promise<GraphEvent[]> {
    const escaped = prefix.replace(/'/g, "''");
    const filter = encodeURIComponent(`startswith(subject,'${escaped}')`);
    let url: string | undefined = `${userPath(email)}/events?$filter=${filter}&$select=id,subject,createdDateTime,attendees&$top=100`;
    const events: GraphEvent[] = [];
    while (url) {
        // eslint-disable-next-line no-await-in-loop
        const page: {value: GraphEvent[]; '@odata.nextLink'?: string} | undefined = await graph('GET', url);
        events.push(...(page?.value ?? []));
        url = page?.['@odata.nextLink'];
    }
    return events;
}

export async function findEventBySubject(email: string, subject: string): Promise<GraphEvent | undefined> {
    const events = await listEventsWithPrefix(email, subject);
    return events.find((event) => event.subject === subject);
}

export async function deleteEvent(email: string, eventId: string) {
    await graph('DELETE', `${userPath(email)}/events/${encodeURIComponent(eventId)}`);
}

// Deletes events whose subject starts with prefix. olderThanMs limits the sweep to
// events created at least that long ago, so a concurrent run keeps its own.
export async function deleteEventsWithPrefix(email: string, prefix: string, olderThanMs = 0) {
    const cutoff = Date.now() - olderThanMs;
    const events = await listEventsWithPrefix(email, prefix);
    const stale = events.filter((event) => new Date(event.createdDateTime).getTime() <= cutoff);
    await Promise.all(stale.map((event) => deleteEvent(email, event.id)));
    return stale.length;
}

// Retries because a freshly created event can take a moment to appear in listings.
export async function waitForEvent(email: string, subject: string, timeoutMs = 30_000): Promise<GraphEvent> {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
        // eslint-disable-next-line no-await-in-loop
        const event = await findEventBySubject(email, subject);
        if (event) {
            return event;
        }
        // eslint-disable-next-line no-await-in-loop
        await new Promise((resolve) => setTimeout(resolve, 2000));
    }
    throw new Error(`Event "${subject}" did not appear in ${email}'s calendar`);
}
