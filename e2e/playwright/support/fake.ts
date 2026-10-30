// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './env';

export const FAKE_URL = process.env.FAKE_URL || 'http://127.0.0.1:8080';

export type FakeEvent = {
    id: string;
    iCalUId: string;
    subject: string;
    start: {dateTime: string; timeZone: string};
    end: {dateTime: string; timeZone: string};
    attendees: Array<{emailAddress: {address: string}; status: {response: string}}>;
    responseStatus: {response: string};
    isCancelled: boolean;
};

export type FakeRequest = {method?: string; path?: string; status?: number; grant?: string; user?: string; kind?: string; changeType?: string};

export type FakeSubscription = {id: string; clientState: string; notificationUrl: string; expirationDateTime: string; creatorId: string};

async function control<T>(method: string, path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${FAKE_URL}/__control/${path}`, {
        method,
        headers: {'content-type': 'application/json'},
        body: body === undefined ? undefined : JSON.stringify(body),
    });
    if (!response.ok && response.status !== 204) {
        throw new Error(`Fake ${method} ${path} failed: ${response.status} ${await response.text()}`);
    }
    return (response.status === 204 ? undefined : await response.json()) as T;
}

export const fake = {
    reset: () => control('POST', 'reset'),
    configure: (config: {accessTokenTtlSeconds?: number; adminConsent?: boolean}) => control('POST', 'config', config),

    createEvent: (email: string, subject: string, start: Date, end: Date, attendees: string[] = []) => control<FakeEvent>('POST', `users/${encodeURIComponent(email)}/events`, {
        subject,
        start: {dateTime: start.toISOString().replace('Z', ''), timeZone: 'UTC'},
        end: {dateTime: end.toISOString().replace('Z', ''), timeZone: 'UTC'},
        attendees: attendees.map((address) => ({emailAddress: {address}})),
    }),
    updateEvent: (email: string, eventId: string, patch: Record<string, unknown>) => control<FakeEvent>('PATCH', `users/${encodeURIComponent(email)}/events/${encodeURIComponent(eventId)}`, patch),
    deleteEvent: (email: string, eventId: string) => control('DELETE', `users/${encodeURIComponent(email)}/events/${encodeURIComponent(eventId)}`),
    events: async (email: string) => (await control<{value: FakeEvent[]}>('GET', `users/${encodeURIComponent(email)}/events`)).value,

    subscriptions: async () => (await control<{value: FakeSubscription[]}>('GET', 'subscriptions')).value,
    notify: (subscriptionId: string, body: {eventId: string; changeType?: string; clientState?: string; soon?: boolean}) => control('POST', `subscriptions/${subscriptionId}/notify`, body),

    requests: async () => (await control<{value: FakeRequest[]}>('GET', 'requests')).value,
    clearRequests: () => control('DELETE', 'requests'),
    violations: async () => (await control<{value: string[]}>('GET', 'violations')).value,

    expireAccessTokens: () => control('POST', 'tokens/expire'),
    revokeRefreshTokens: () => control('POST', 'tokens/revoke'),
};

export const createEvent = fake.createEvent;

export async function waitForEvent(email: string, subject: string, timeoutMs = 15_000): Promise<FakeEvent> {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
        // eslint-disable-next-line no-await-in-loop
        const found = (await fake.events(email)).find((event) => event.subject === subject);
        if (found) {
            return found;
        }
        // eslint-disable-next-line no-await-in-loop
        await new Promise((resolve) => setTimeout(resolve, 500));
    }
    throw new Error(`Event "${subject}" did not appear in ${email}'s calendar`);
}

export async function deleteEventsWithPrefix(email: string, prefix: string) {
    const matching = (await fake.events(email)).filter((event) => event.subject.startsWith(prefix));
    await Promise.all(matching.map((event) => fake.deleteEvent(email, event.id)));
}
