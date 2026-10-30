// A fake of Microsoft Entra ID sign-in and the Microsoft Graph calls the plugin makes.
// It listens on HTTPS 443 for Mattermost (reached through the Compose network aliases
// login.microsoftonline.com and graph.microsoft.com) and on HTTP 8080 for the browser
// and the test control API. Uses only Node built-ins, with type annotations that Node strips.
import crypto from 'node:crypto';
import fs from 'node:fs';
import http from 'node:http';
import https from 'node:https';

import {missingFields} from './contract.ts';

const CLIENT_ID = process.env.FAKE_CLIENT_ID || 'fake-client-id';
const CLIENT_SECRET = process.env.FAKE_CLIENT_SECRET || 'fake-client-secret';
const TENANT_ID = process.env.FAKE_TENANT_ID || 'fake-tenant';
const REDIRECT_URIS = (process.env.FAKE_REDIRECT_URIS || 'http://localhost:8065/plugins/com.mattermost.mscalendar/oauth2/complete').split(',');
const MATTERMOST_URL = process.env.FAKE_MATTERMOST_URL || 'http://mattermost:8065';
const MAX_SUBSCRIPTION_MINUTES = 4230;

type Json = Record<string, any>;
type Result = {status: number; body?: Json | string; headers?: Record<string, string>};

const USERS = [
    {email: 'alice@contoso.test', name: 'Alice Fake', password: 'Alice-Fake-Pass1'},
    {email: 'bob@contoso.test', name: 'Bob Fake', password: 'Bob-Fake-Pass1'},
].map((u) => ({...u, id: crypto.createHash('sha1').update(u.email).digest('hex').replace(/^(.{8})(.{4})(.{4})(.{4})(.{12}).*$/, '$1-$2-$3-$4-$5')}));
type User = typeof USERS[number];

type AccessToken = {user: User | null; expires: number};
type Code = {user: User; redirectUri: string; scope: string; expires: number};
type Pending = {query: URLSearchParams; user?: User};
type Subscription = Json & {id: string; creatorEmail: string};

const state = {
    accessTokens: new Map<string, AccessToken>(),
    refreshTokens: new Map<string, {user: User; scope: string}>(),
    codes: new Map<string, Code>(),
    pending: new Map<string, Pending>(),
    events: new Map<string, Json[]>(),
    subscriptions: new Map<string, Subscription>(),
    requests: [] as Json[],
    violations: [] as string[],
    config: {accessTokenTtlSeconds: 3600, adminConsent: true},
};

function reset() {
    state.accessTokens.clear();
    state.refreshTokens.clear();
    state.codes.clear();
    state.pending.clear();
    state.events.clear();
    state.subscriptions.clear();
    state.requests = [];
    state.violations = [];
    state.config = {accessTokenTtlSeconds: 3600, adminConsent: true};
}

const uid = () => crypto.randomUUID();
const token = (prefix: string) => `${prefix}.${crypto.randomBytes(24).toString('base64url')}`;
const html = (s: string) => s.replace(/[&<>"]/g, (c) => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;'}[c] as string));

function findUser(idOrEmail: string): User | undefined {
    const key = decodeURIComponent(idOrEmail).toLowerCase();
    return USERS.find((u) => u.id === key || u.email === key);
}

function validate(contractKey: string, value: unknown, where: string) {
    const missing = missingFields(contractKey, value);
    if (missing.length > 0) {
        state.violations.push(`${where}: missing ${missing.join(', ')}`);
    }
}

// ---- Entra ID sign-in pages ----

function page(body: string): Result {
    return {
        status: 200,
        headers: {'content-type': 'text/html; charset=utf-8'},
        body: `<!doctype html><html><head><title>Sign in to your account</title></head><body>${body}</body></html>`,
    };
}

function errorPage(status: number, message: string): Result {
    return {status, headers: {'content-type': 'text/html; charset=utf-8'}, body: `<!doctype html><html><body><h1>Sign in error</h1><p id="service_exception_message">${html(message)}</p></body></html>`};
}

function hidden(sid: string) {
    return `<input type="hidden" name="sid" value="${sid}">`;
}

function emailPage(tenant: string, sid: string, error = ''): Result {
    return page(`<form method="post" action="/${tenant}/login/email">${hidden(sid)}
        <h1>Sign in</h1>${error ? `<div id="usernameError">${html(error)}</div>` : ''}
        <input type="email" name="loginfmt" placeholder="Email, phone, or Skype">
        <input type="submit" id="idSIButton9" value="Next"></form>`);
}

function passwordPage(tenant: string, sid: string, email: string, error = ''): Result {
    return page(`<form method="post" action="/${tenant}/login/password">${hidden(sid)}
        <h1>Enter password</h1><div id="displayName">${html(email)}</div>
        ${error ? `<div id="passwordError">${html(error)}</div>` : ''}
        <input type="password" name="passwd" placeholder="Password">
        <input type="submit" id="idSIButton9" value="Sign in"></form>`);
}

function consentPage(tenant: string, sid: string): Result {
    return page(`<form method="post" action="/${tenant}/login/consent">${hidden(sid)}
        <h1>Permissions requested</h1><p>Let this app access your info and calendars.</p>
        <button type="submit" id="idSIButton9" name="action" value="accept">Accept</button>
        <button type="submit" id="cancelConsent" name="action" value="deny">Cancel</button></form>`);
}

function kmsiPage(tenant: string, sid: string): Result {
    return page(`<form method="post" action="/${tenant}/login/kmsi">${hidden(sid)}
        <h1>Stay signed in?</h1>
        <button type="submit" id="idSIButton9" name="action" value="yes">Yes</button>
        <button type="submit" id="idBtn_Back" name="action" value="no">No</button></form>`);
}

function redirect(location: string): Result {
    return {status: 302, headers: {location}, body: ''};
}

function redirectWith(redirectUri: string, params: Record<string, string>): Result {
    const url = new URL(redirectUri);
    for (const [key, value] of Object.entries(params)) {
        url.searchParams.set(key, value);
    }
    return redirect(url.toString());
}

function authorize(tenant: string, query: URLSearchParams): Result {
    const clientId = query.get('client_id');
    const redirectUri = query.get('redirect_uri') || '';
    if (clientId !== CLIENT_ID) {
        return errorPage(400, `AADSTS700016: Application with identifier '${clientId}' was not found in the directory.`);
    }
    if (!REDIRECT_URIS.includes(redirectUri)) {
        return errorPage(400, `AADSTS50011: The redirect URI '${redirectUri}' specified in the request does not match the redirect URIs configured for the application '${clientId}'.`);
    }
    if (query.get('response_type') !== 'code') {
        return errorPage(400, 'AADSTS700054: response_type \'code\' is not enabled for the application.');
    }

    const sid = uid();
    state.pending.set(sid, {query});
    return emailPage(tenant, sid);
}

function finishSignIn(pending: Pending): Result {
    const redirectUri = pending.query.get('redirect_uri') as string;
    const code = token('code');
    state.codes.set(code, {
        user: pending.user as User,
        redirectUri,
        scope: pending.query.get('scope') || '',
        expires: Date.now() + (10 * 60 * 1000),
    });
    return redirectWith(redirectUri, {code, state: pending.query.get('state') || '', session_state: uid()});
}

function loginStep(tenant: string, step: string, form: URLSearchParams): Result {
    const sid = form.get('sid') || '';
    const pending = state.pending.get(sid);
    if (!pending) {
        return errorPage(400, 'AADSTS90016: The sign-in session was not found. Start again.');
    }

    switch (step) {
    case 'email': {
        const user = findUser(form.get('loginfmt') || '');
        if (!user) {
            return emailPage(tenant, sid, "We couldn't find an account with that username.");
        }
        pending.user = user;
        return passwordPage(tenant, sid, user.email);
    }
    case 'password': {
        const user = pending.user;
        if (!user || form.get('passwd') !== user.password) {
            return passwordPage(tenant, sid, user?.email || '', 'Your account or password is incorrect.');
        }
        if (pending.query.get('prompt') === 'consent' || !state.config.adminConsent) {
            return consentPage(tenant, sid);
        }
        return kmsiPage(tenant, sid);
    }
    case 'consent':
        if (form.get('action') === 'deny') {
            return redirectWith(pending.query.get('redirect_uri') as string, {
                error: 'access_denied',
                error_description: 'AADSTS65004: The user declined to consent to access the app.',
                state: pending.query.get('state') || '',
            });
        }
        return kmsiPage(tenant, sid);
    case 'kmsi':
        state.pending.delete(sid);
        return finishSignIn(pending);
    default:
        return errorPage(404, 'Not found');
    }
}

// ---- Token endpoint ----

function tokenError(status: number, error: string, code: number, description: string): Result {
    return {
        status,
        body: {
            error,
            error_description: `AADSTS${code}: ${description}\r\nTrace ID: ${uid()}\r\nCorrelation ID: ${uid()}\r\nTimestamp: ${new Date().toISOString()}`,
            error_codes: [code],
            timestamp: new Date().toISOString(),
            trace_id: uid(),
            correlation_id: uid(),
        },
    };
}

function issueTokens(user: User | null, scope: string, withRefresh: boolean): Result {
    const accessToken = token('at');
    const ttl = state.config.accessTokenTtlSeconds;
    state.accessTokens.set(accessToken, {user, expires: Date.now() + (ttl * 1000)});

    const body: Json = {token_type: 'Bearer', scope, expires_in: ttl, ext_expires_in: ttl, access_token: accessToken};
    if (withRefresh && user) {
        body.refresh_token = token('rt');
        state.refreshTokens.set(body.refresh_token, {user, scope});
    }
    validate('token', body, 'token endpoint');
    return {status: 200, body};
}

function tokenEndpoint(form: URLSearchParams, authHeader: string): Result {
    let clientId = form.get('client_id');
    let clientSecret = form.get('client_secret');
    if (authHeader.startsWith('Basic ')) {
        const [id, secret] = Buffer.from(authHeader.slice(6), 'base64').toString().split(':');
        clientId = decodeURIComponent(id);
        clientSecret = decodeURIComponent(secret || '');
    }
    if (clientId !== CLIENT_ID) {
        return tokenError(401, 'invalid_client', 700016, `Application with identifier '${clientId}' was not found in the directory.`);
    }
    if (clientSecret !== CLIENT_SECRET) {
        return tokenError(401, 'invalid_client', 7000215, 'Invalid client secret provided. Ensure the secret being sent in the request is the client secret value, not the client secret ID.');
    }

    const grant = form.get('grant_type');
    switch (grant) {
    case 'authorization_code': {
        const code = state.codes.get(form.get('code') || '');
        state.codes.delete(form.get('code') || '');
        if (!code || code.expires < Date.now()) {
            return tokenError(400, 'invalid_grant', 70000, 'The provided value for the \'code\' parameter is not valid. The code has expired.');
        }
        if (code.redirectUri !== form.get('redirect_uri')) {
            return tokenError(400, 'invalid_grant', 50011, 'The redirect URI does not match the one used to obtain the code.');
        }
        return issueTokens(code.user, code.scope, true);
    }
    case 'refresh_token': {
        const refresh = state.refreshTokens.get(form.get('refresh_token') || '');
        if (!refresh) {
            return tokenError(400, 'invalid_grant', 700082, 'The refresh token has expired due to inactivity. The token was issued on 2026-01-01T00:00:00.0000000Z and was inactive for 90.00:00:00.');
        }
        return issueTokens(refresh.user, refresh.scope, true);
    }
    case 'client_credentials':
        if (form.get('scope') !== 'https://graph.microsoft.com/.default') {
            return tokenError(400, 'invalid_scope', 70011, 'The provided request must include a \'scope\' input parameter. The provided value for the input parameter \'scope\' is not valid.');
        }
        return issueTokens(null, 'https://graph.microsoft.com/.default', false);
    default:
        return tokenError(400, 'unsupported_grant_type', 70003, `The app requested an unsupported grant type '${grant}'.`);
    }
}

// ---- Graph ----

function graphError(status: number, code: string, message: string): Result {
    const body = {error: {code, message, innerError: {date: new Date().toISOString(), 'request-id': uid(), 'client-request-id': uid()}}};
    validate('graphError', body, 'graph error');
    return {status, body};
}

type Auth = {user: User | null};

function authenticate(header: string): Auth | Result {
    const bearer = header.startsWith('Bearer ') ? header.slice(7) : '';
    if (!bearer) {
        return graphError(401, 'InvalidAuthenticationToken', 'Access token is empty.');
    }
    const found = state.accessTokens.get(bearer);
    if (!found || found.expires < Date.now()) {
        return graphError(401, 'InvalidAuthenticationToken', 'Access token has expired or is not yet valid.');
    }
    return {user: found.user};
}

function mailbox(user: User): Json[] {
    if (!state.events.has(user.email)) {
        state.events.set(user.email, []);
    }
    return state.events.get(user.email) as Json[];
}

function toUtc(dateTime: string, timeZone: string): Date {
    if (/[zZ]$|[+-]\d\d:\d\d$/.test(dateTime)) {
        return new Date(dateTime);
    }
    const asUtc = new Date(`${dateTime.slice(0, 23)}Z`);
    if (!timeZone || timeZone === 'UTC') {
        return asUtc;
    }
    try {
        const wall = new Intl.DateTimeFormat('en-US', {timeZone, hourCycle: 'h23', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit'}).formatToParts(asUtc);
        const get = (type: string) => Number(wall.find((p) => p.type === type)?.value);
        const shown = Date.UTC(get('year'), get('month') - 1, get('day'), get('hour'), get('minute'), get('second'));
        return new Date(asUtc.getTime() - (shown - asUtc.getTime()));
    } catch {
        return asUtc;
    }
}

const graphDate = (d: Date) => ({dateTime: `${d.toISOString().slice(0, 23)}0000`.replace('Z', ''), timeZone: 'UTC'});

function newEvent(owner: User, input: Json, iCalUId: string, organizer: User): Json {
    const start = toUtc(input.start.dateTime, input.start.timeZone);
    const end = toUtc(input.end.dateTime, input.end.timeZone);
    const attendees = (input.attendees || []).map((a: Json) => ({
        type: a.type || 'required',
        status: {response: 'none', time: '0001-01-01T00:00:00Z'},
        emailAddress: {address: a.emailAddress.address, name: findUser(a.emailAddress.address)?.name || a.emailAddress.address},
    }));
    const id = `AAMk${crypto.randomBytes(24).toString('base64url')}`;
    const isOrganizer = owner.email === organizer.email;
    const content = input.body?.content || '';
    return {
        id,
        iCalUId,
        subject: input.subject || '',
        body: {contentType: input.body?.contentType || 'text', content},
        bodyPreview: content.slice(0, 255),
        start: graphDate(start),
        end: graphDate(end),
        location: {displayName: input.location?.displayName || ''},
        attendees,
        organizer: {emailAddress: {address: organizer.email, name: organizer.name}},
        isOrganizer,
        isCancelled: false,
        isAllDay: Boolean(input.isAllDay),
        showAs: input.showAs || 'busy',
        importance: 'normal',
        reminderMinutesBeforeStart: 15,
        responseRequested: true,
        responseStatus: {response: isOrganizer ? 'organizer' : 'none', time: '0001-01-01T00:00:00Z'},
        webLink: `https://outlook.office.com/calendar/item/${encodeURIComponent(id)}`,
        createdDateTime: new Date().toISOString(),
        lastModifiedDateTime: new Date().toISOString(),
    };
}

function eventBounds(event: Json) {
    return {start: toUtc(event.start.dateTime, event.start.timeZone).getTime(), end: toUtc(event.end.dateTime, event.end.timeZone).getTime()};
}

// Creates the organizer's copy and delivers an invitation copy to every attendee that has a mailbox.
function createEvent(organizer: User, input: Json): Json {
    const iCalUId = uid().replace(/-/g, '').toUpperCase();
    const own = newEvent(organizer, input, iCalUId, organizer);
    mailbox(organizer).push(own);
    notifyChange(organizer, 'created', own);

    for (const attendee of own.attendees) {
        const invitee = findUser(attendee.emailAddress.address);
        if (invitee && invitee.email !== organizer.email) {
            const copy = newEvent(invitee, input, iCalUId, organizer);
            mailbox(invitee).push(copy);
            notifyChange(invitee, 'created', copy);
        }
    }
    return own;
}

function respond(user: User, eventId: string, response: string): boolean {
    const mine = mailbox(user).find((e) => e.id === eventId);
    if (!mine) {
        return false;
    }
    mine.responseStatus = {response, time: new Date().toISOString()};
    for (const [email, events] of state.events) {
        for (const event of events) {
            if (event.iCalUId === mine.iCalUId && event.isOrganizer) {
                const attendee = event.attendees.find((a: Json) => a.emailAddress.address === user.email);
                if (attendee) {
                    attendee.status = {response, time: new Date().toISOString()};
                    notifyChange(findUser(email) as User, 'updated', event);
                }
            }
        }
    }
    notifyChange(user, 'updated', mine);
    return true;
}

function cancelEvent(user: User, eventId: string): boolean {
    const events = mailbox(user);
    const index = events.findIndex((e) => e.id === eventId);
    if (index === -1) {
        return false;
    }
    const [removed] = events.splice(index, 1);
    notifyChange(user, 'deleted', removed);
    return true;
}

function notFound(): Result {
    return graphError(404, 'ErrorItemNotFound', 'The specified object was not found in the store.');
}

function resolveUser(auth: Auth, ref: string): User | Result {
    if (ref === 'me') {
        if (!auth.user) {
            return graphError(400, 'BadRequest', '/me request is only valid with delegated authentication flow.');
        }
        return auth.user;
    }
    const user = findUser(ref);
    if (!user) {
        return graphError(404, 'ErrorInvalidUser', 'The requested user is invalid.');
    }
    if (auth.user && auth.user.email !== user.email) {
        return graphError(403, 'ErrorAccessDenied', 'Access is denied. Check credentials and try again.');
    }
    return user;
}

const isResult = (value: unknown): value is Result => typeof value === 'object' && value !== null && 'status' in value;

function graphRoute(method: string, pathname: string, query: URLSearchParams, body: Json, auth: Auth): Result {
    const path = pathname.replace(/^\/v1\.0/, '').replace(/\/+$/, '');
    const segments = path.split('/').filter(Boolean);
    const root = segments[0]?.toLowerCase();

    if (root === '$batch' && method === 'POST') {
        return batch(body, auth);
    }
    if (root === 'subscriptions') {
        return subscriptionsRoute(method, segments, body, auth);
    }

    if (root !== 'me' && root !== 'users') {
        return graphError(404, 'ResourceNotFound', `Resource not found for the segment '${segments[0] || ''}'.`);
    }
    return userRoute(method, pathname, query, body, auth);
}

function userRoute(method: string, pathname: string, query: URLSearchParams, body: Json, auth: Auth): Result {
    const segments = pathname.replace(/^\/v1\.0/, '').split('/').filter(Boolean);
    const ref = segments[0].toLowerCase() === 'me' ? 'me' : segments[1];
    const rest = segments.slice(segments[0].toLowerCase() === 'me' ? 1 : 2).map((s) => s.toLowerCase());
    const original = segments.slice(segments[0].toLowerCase() === 'me' ? 1 : 2);

    const user = resolveUser(auth, ref);
    if (isResult(user)) {
        return user;
    }

    if (rest.length === 0 && method === 'GET') {
        const profile = {id: user.id, displayName: user.name, userPrincipalName: user.email, mail: user.email, givenName: user.name.split(' ')[0], surname: user.name.split(' ')[1]};
        validate('user', profile, 'GET user');
        return {status: 200, body: profile};
    }

    if (rest[0] === 'mailboxsettings' && method === 'GET') {
        const settings = {timeZone: 'UTC', workingHours: {daysOfWeek: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'], startTime: '08:00:00.0000000', endTime: '17:00:00.0000000', timeZone: {name: 'UTC'}}};
        validate('mailboxSettings', settings, 'GET mailboxSettings');
        return {status: 200, body: settings};
    }

    if (rest[0] === 'calendars') {
        return calendarsRoute(method, user, original, body);
    }

    if (rest[0] === 'getschedule' && method === 'POST') {
        const interval = Number(body.availabilityViewInterval || 30);
        const from = toUtc(body.startTime.dateTime, body.startTime.timeZone).getTime();
        const to = toUtc(body.endTime.dateTime, body.endTime.timeZone).getTime();
        const slots = Math.max(1, Math.ceil((to - from) / (interval * 60000)));
        const value = (body.schedules || []).map((scheduleId: string) => {
            const owner = findUser(scheduleId);
            const events = owner ? mailbox(owner) : [];
            const view = Array.from({length: slots}, (_, i) => {
                const slotStart = from + (i * interval * 60000);
                const busy = events.some((e) => {
                    const b = eventBounds(e);
                    return !e.isCancelled && b.start < slotStart + (interval * 60000) && b.end > slotStart;
                });
                return busy ? '2' : '0';
            }).join('');
            const info = {scheduleId, availabilityView: view, scheduleItems: []};
            validate('schedule', info, 'getSchedule');
            return info;
        });
        return {status: 200, body: {value}};
    }

    if (rest[0] === 'findmeetingtimes' && method === 'POST') {
        return {status: 200, body: {meetingTimeSuggestions: [], emptySuggestionsReason: 'AttendeesUnavailable'}};
    }

    if (rest[0] === 'calendarview' && method === 'GET') {
        const from = new Date(query.get('startDateTime') || 0).getTime();
        const to = new Date(query.get('endDateTime') || 0).getTime();
        const value = mailbox(user).filter((e) => {
            const b = eventBounds(e);
            return b.start < to && b.end > from;
        });
        value.forEach((e) => validate('event', e, 'calendarView event'));
        return {status: 200, body: {value}};
    }

    if (rest[0] === 'events') {
        return eventsRoute(method, user, original, body);
    }

    return graphError(404, 'ResourceNotFound', `Resource not found for the segment '${original[0] || ''}'.`);
}

function calendarsRoute(method: string, user: User, segments: string[], body: Json): Result {
    const calendars = [{id: `cal-${user.id}`, name: 'Calendar', owner: {name: user.name, address: user.email}, canEdit: true}];
    if (segments.length === 1 && method === 'GET') {
        calendars.forEach((c) => validate('calendar', c, 'GET calendars'));
        return {status: 200, body: {value: calendars}};
    }
    if (segments.length === 1 && method === 'POST') {
        const created = {id: `cal-${uid()}`, name: body.name || '', owner: {name: user.name, address: user.email}, canEdit: true};
        validate('calendar', created, 'POST calendars');
        return {status: 201, body: created};
    }
    if (segments.length === 2 && method === 'DELETE') {
        return {status: 204};
    }
    return graphError(405, 'MethodNotAllowed', 'The HTTP method is not allowed for this resource.');
}

function eventsRoute(method: string, user: User, segments: string[], body: Json): Result {
    if (segments.length === 1 && method === 'POST') {
        const problems = ['subject', 'start.dateTime', 'start.timeZone', 'end.dateTime', 'end.timeZone'].filter((p) => p.split('.').reduce((v: any, k) => v?.[k], body) === undefined);
        if (problems.length > 0) {
            state.violations.push(`create event request: missing ${problems.join(', ')}`);
            return graphError(400, 'ErrorInvalidRequest', `The request is missing ${problems.join(', ')}.`);
        }
        const created = createEvent(user, body);
        validate('event', created, 'POST event');
        return {status: 201, body: created};
    }

    const event = mailbox(user).find((e) => e.id === decodeURIComponent(segments[1] || ''));
    if (!event) {
        return notFound();
    }
    const action = (segments[2] || '').toLowerCase();

    if (segments.length === 2 && method === 'GET') {
        validate('event', event, 'GET event');
        return {status: 200, body: event};
    }
    if (segments.length === 2 && method === 'DELETE') {
        cancelEvent(user, event.id);
        return {status: 204};
    }
    if (segments.length === 2 && method === 'PATCH') {
        Object.assign(event, body, {lastModifiedDateTime: new Date().toISOString()});
        notifyChange(user, 'updated', event);
        return {status: 200, body: event};
    }
    const responses: Record<string, string> = {accept: 'accepted', decline: 'declined', tentativelyaccept: 'tentativelyAccepted'};
    if (method === 'POST' && responses[action]) {
        respond(user, event.id, responses[action]);
        return {status: 202};
    }
    return graphError(405, 'MethodNotAllowed', 'The HTTP method is not allowed for this resource.');
}

function batch(body: Json, auth: Auth): Result {
    const requests: Json[] = body.requests || [];
    if (requests.length > 20) {
        return graphError(400, 'BadRequest', 'Batch request limit is 20.');
    }
    const responses = requests.map((r) => {
        const url = new URL(`http://batch${r.url.startsWith('/') ? '' : '/'}${r.url}`);
        const result = graphRoute((r.method || 'GET').toUpperCase(), `/v1.0${url.pathname}`, url.searchParams, r.body || {}, auth);
        return {id: r.id, status: result.status, headers: {'Content-Type': 'application/json'}, body: result.body};
    });
    return {status: 200, body: {responses}};
}

// ---- Subscriptions and change notifications ----

function subscriptionsRoute(method: string, segments: string[], body: Json, auth: Auth): Result {
    if (!auth.user) {
        return graphError(403, 'ExtensionError', 'Operation: Create; Exception: [Status Code: Forbidden; Reason: Application permissions are not supported for this resource.]');
    }
    const user = auth.user;

    if (segments.length === 1 && method === 'GET') {
        return {status: 200, body: {value: [...state.subscriptions.values()].filter((s) => s.creatorEmail === user.email).map(publicSubscription)}};
    }
    if (segments.length === 1 && method === 'POST') {
        return createSubscription(user, body);
    }

    const sub = state.subscriptions.get(segments[1]);
    if (!sub || sub.creatorEmail !== user.email) {
        return graphError(404, 'ResourceNotFound', 'The object was not found.');
    }
    if (method === 'DELETE') {
        state.subscriptions.delete(sub.id);
        return {status: 204};
    }
    if (method === 'PATCH') {
        const minutes = (new Date(body.expirationDateTime).getTime() - Date.now()) / 60000;
        if (!(minutes > 0) || minutes > MAX_SUBSCRIPTION_MINUTES) {
            return graphError(400, 'ValidationError', `Subscription expiration can only be ${MAX_SUBSCRIPTION_MINUTES} minutes in the future.`);
        }
        sub.expirationDateTime = body.expirationDateTime;
        return {status: 200, body: publicSubscription(sub)};
    }
    return graphError(405, 'MethodNotAllowed', 'The HTTP method is not allowed for this resource.');
}

function publicSubscription(sub: Subscription): Json {
    const {creatorEmail: _creator, ...rest} = sub;
    return rest;
}

// Mattermost builds the notification URL from its SiteURL (localhost), which the fake cannot reach.
function deliveryUrl(notificationUrl: string): string {
    const url = new URL(notificationUrl);
    return `${MATTERMOST_URL}${url.pathname}${url.search}`;
}

function createSubscription(user: User, body: Json): Result {
    const missing = missingFields('createSubscriptionRequest', body);
    if (missing.length > 0) {
        state.violations.push(`create subscription request: missing ${missing.join(', ')}`);
        return graphError(400, 'ValidationError', `Missing ${missing.join(', ')}.`);
    }
    if (!['me/events', `users/${user.id}/events`].includes(String(body.resource).replace(/^\//, '').toLowerCase())) {
        return graphError(400, 'ValidationError', `Subscription resource '${body.resource}' is not supported by the fake.`);
    }
    const minutes = (new Date(body.expirationDateTime).getTime() - Date.now()) / 60000;
    if (!(minutes > 0) || minutes > MAX_SUBSCRIPTION_MINUTES) {
        return graphError(400, 'ValidationError', `Subscription expiration can only be ${MAX_SUBSCRIPTION_MINUTES} minutes in the future.`);
    }
    if (String(body.clientState || '').length > 128) {
        return graphError(400, 'ValidationError', 'The client state must not exceed 128 characters.');
    }

    const sub: Subscription = {
        id: uid(),
        resource: body.resource,
        applicationId: CLIENT_ID,
        changeType: body.changeType,
        clientState: body.clientState || '',
        notificationUrl: body.notificationUrl,
        expirationDateTime: body.expirationDateTime,
        creatorId: user.id,
        creatorEmail: user.email,
    };
    validate('subscription', sub, 'POST subscription');

    // Graph answers the create call only after the endpoint echoes a validation token. This
    // fake does so asynchronously through a deferred response.
    return {status: -1, body: {sub}};
}

async function validateEndpoint(notificationUrl: string): Promise<boolean> {
    const token = uid();
    try {
        const response = await fetch(`${deliveryUrl(notificationUrl)}?validationToken=${token}`, {method: 'POST', signal: AbortSignal.timeout(10000)});
        return response.status === 200 && (await response.text()) === token;
    } catch {
        return false;
    }
}

function notifyChange(owner: User, changeType: string, event: Json, extra: Json = {}) {
    for (const sub of state.subscriptions.values()) {
        if (sub.creatorEmail === owner.email && String(sub.changeType).split(',').includes(changeType)) {
            deliver(sub, changeType, event, extra);
        }
    }
}

function deliver(sub: Subscription, changeType: string, event: Json, extra: Json = {}) {
    const notification = {
        subscriptionId: sub.id,
        subscriptionExpirationDateTime: extra.subscriptionExpirationDateTime || sub.expirationDateTime,
        changeType,
        resource: `Users/${sub.creatorId}/Events/${event.id}`,
        resourceData: {'@odata.type': '#Microsoft.Graph.Event', '@odata.id': `Users/${sub.creatorId}/Events/${event.id}`, id: event.id},
        clientState: extra.clientState === undefined ? sub.clientState : extra.clientState,
        tenantId: TENANT_ID,
    };
    validate('notification', notification, 'change notification');
    state.requests.push({kind: 'notification', subscriptionId: sub.id, changeType, eventId: event.id});

    setTimeout(() => {
        fetch(deliveryUrl(sub.notificationUrl), {
            method: 'POST',
            headers: {'content-type': 'application/json'},
            body: JSON.stringify({value: [notification]}),
        }).catch((err) => console.error('notification delivery failed', err));
    }, 200);
}

// ---- HTTP plumbing ----

async function readBody(req: http.IncomingMessage): Promise<string> {
    const chunks: Buffer[] = [];
    for await (const chunk of req) {
        chunks.push(chunk as Buffer);
    }
    return Buffer.concat(chunks).toString();
}

function send(res: http.ServerResponse, result: Result) {
    const isString = typeof result.body === 'string';
    const payload = result.body === undefined ? '' : isString ? result.body as string : JSON.stringify(result.body);
    res.writeHead(result.status, {
        'content-type': isString ? 'text/plain' : 'application/json',
        ...result.headers,
        'content-length': Buffer.byteLength(payload),
    });
    res.end(payload);
}

async function control(method: string, pathname: string, body: Json): Promise<Result> {
    const segments = pathname.split('/').filter(Boolean).slice(1);
    const [section, ref, sub] = segments;

    if (section === 'health') {
        return {status: 200, body: {ok: true}};
    }
    if (section === 'reset' && method === 'POST') {
        reset();
        return {status: 200, body: {ok: true}};
    }
    if (section === 'config' && method === 'POST') {
        Object.assign(state.config, body);
        return {status: 200, body: state.config};
    }
    if (section === 'requests') {
        if (method === 'DELETE') {
            state.requests = [];
            return {status: 200, body: {ok: true}};
        }
        return {status: 200, body: {value: state.requests}};
    }
    if (section === 'violations') {
        return {status: 200, body: {value: state.violations}};
    }
    if (section === 'tokens' && method === 'POST') {
        if (ref === 'expire') {
            state.accessTokens.forEach((t) => {
                t.expires = 0;
            });
        } else if (ref === 'revoke') {
            state.refreshTokens.clear();
        }
        return {status: 200, body: {ok: true}};
    }
    if (section === 'subscriptions' && !ref && method === 'GET') {
        return {status: 200, body: {value: [...state.subscriptions.values()].map(publicSubscription)}};
    }
    if (section === 'subscriptions' && ref && sub === 'notify' && method === 'POST') {
        const subscription = state.subscriptions.get(ref);
        const owner = subscription && findUser(subscription.creatorEmail);
        const event = owner && mailbox(owner).find((e) => e.id === body.eventId);
        if (!subscription || !event) {
            return {status: 404, body: {error: 'subscription or event not found'}};
        }
        deliver(subscription, body.changeType || 'updated', event, {
            clientState: body.clientState,
            subscriptionExpirationDateTime: body.soon ? new Date(Date.now() + (60 * 60 * 1000)).toISOString() : undefined,
        });
        return {status: 202, body: {ok: true}};
    }
    if (section === 'users' && sub === 'events') {
        const user = findUser(ref);
        if (!user) {
            return {status: 404, body: {error: 'unknown user'}};
        }
        if (method === 'POST') {
            return {status: 201, body: createEvent(user, body)};
        }
        if (method === 'GET') {
            return {status: 200, body: {value: mailbox(user)}};
        }
        if (method === 'PATCH') {
            const event = mailbox(user).find((e) => e.id === segments[3]);
            if (!event) {
                return {status: 404, body: {error: 'not found'}};
            }
            Object.assign(event, body, {lastModifiedDateTime: new Date().toISOString()});
            notifyChange(user, 'updated', event);
            return {status: 200, body: event};
        }
        if (method === 'DELETE') {
            return cancelEvent(user, segments[3]) ? {status: 204} : {status: 404, body: {error: 'not found'}};
        }
    }
    return {status: 404, body: {error: 'unknown control route'}};
}

async function handle(req: http.IncomingMessage, res: http.ServerResponse) {
    const url = new URL(req.url || '/', 'http://fake');
    const method = req.method || 'GET';
    const raw = await readBody(req);
    const contentType = req.headers['content-type'] || '';
    const isJson = contentType.includes('json');
    const form = new URLSearchParams(isJson ? '' : raw);
    let json: Json = {};
    if (isJson && raw) {
        try {
            json = JSON.parse(raw);
        } catch {
            json = {};
        }
    }

    let result: Result;
    const logEntry: Json = {method, path: url.pathname, time: new Date().toISOString()};

    if (url.pathname.startsWith('/__control/')) {
        result = await control(method, url.pathname, json);
        return send(res, result);
    }

    const match = url.pathname.match(/^\/([^/]+)\/(?:oauth2\/v2\.0\/(authorize|token)|login\/(email|password|consent|kmsi))$/);
    if (match) {
        const [, tenant, oauthStep, loginStepName] = match;
        if (oauthStep === 'authorize' && method === 'GET') {
            result = authorize(tenant, url.searchParams);
        } else if (oauthStep === 'token' && method === 'POST') {
            logEntry.grant = form.get('grant_type');
            result = tokenEndpoint(form, String(req.headers.authorization || ''));
        } else if (loginStepName && method === 'POST') {
            result = loginStep(tenant, loginStepName, form);
        } else {
            result = errorPage(405, 'Method not allowed');
        }
    } else if (url.pathname.startsWith('/v1.0/')) {
        const auth = authenticate(String(req.headers.authorization || ''));
        if (isResult(auth)) {
            result = auth;
        } else {
            logEntry.user = auth.user?.email || 'app-only';
            result = graphRoute(method, url.pathname, url.searchParams, json, auth);
            if (result.status === -1) {
                const sub = (result.body as Json).sub as Subscription;
                if (await validateEndpoint(sub.notificationUrl)) {
                    state.subscriptions.set(sub.id, sub);
                    result = {status: 201, body: publicSubscription(sub)};
                } else {
                    result = graphError(400, 'ValidationError', 'Subscription validation request failed. Notification endpoint must respond with 200 OK to validation request.');
                }
            }
        }
    } else {
        result = {status: 404, body: 'Not found'};
    }

    logEntry.status = result.status;
    state.requests.push(logEntry);
    send(res, result);
}

const onRequest = (req: http.IncomingMessage, res: http.ServerResponse) => {
    handle(req, res).catch((err) => {
        console.error(err);
        send(res, {status: 500, body: {error: String(err)}});
    });
};

http.createServer(onRequest).listen(8080, () => console.log('fake microsoft: http on 8080'));

const certDir = process.env.FAKE_CERT_DIR || '/certs';
if (fs.existsSync(`${certDir}/fake.crt`)) {
    https.createServer({cert: fs.readFileSync(`${certDir}/fake.crt`), key: fs.readFileSync(`${certDir}/fake.key`)}, onRequest).listen(443, () => console.log('fake microsoft: https on 443'));
}
