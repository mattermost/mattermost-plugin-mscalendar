# Microsoft Calendar plugin E2E tests

Playwright suite that drives the plugin in a real browser against a disposable Mattermost server. Microsoft Entra ID and Microsoft Graph cannot run in a container, so the suite talks to a **fake** that implements the parts of both the plugin uses. A nightly **contract check** keeps the fake honest against a real tenant.

Built on [`@mattermost/playwright-lib`](https://www.npmjs.com/package/@mattermost/playwright-lib), so no `mattermost` monorepo checkout is needed. No secrets are required to run the suite, so it works for forks and Cursor Cloud Agents.

## How it works

```
 browser ──login.microsoftonline.com (rewritten by Playwright)──┐
                                                                ▼
 Mattermost + plugin ──https://login.microsoftonline.com ──▶  fake-microsoft
                     ──https://graph.microsoft.com ───────▶  (Node, no dependencies)
                     ◀── change notifications ─────────────
```

- `fake-microsoft/server.ts` serves OAuth v2 (authorize with login, consent and "stay signed in" pages; the authorization-code, refresh-token and client-credentials grants; AADSTS errors) and the Graph v1.0 calls the plugin makes (profile, mailbox settings, calendars, calendar view, events and responses, `getSchedule`, `findMeetingTimes`, `$batch`, subscriptions with the validation handshake, `clientState` and expiry limits). It delivers change notifications to Mattermost like Microsoft does.
- The plugin hardcodes the Microsoft hosts, so Docker Compose gives the fake the network aliases `login.microsoftonline.com` and `graph.microsoft.com`. A throwaway CA, created by `fake-microsoft/gen-certs.sh`, is trusted by Mattermost through `SSL_CERT_FILE`. The browser is redirected to the fake by a Playwright route (`routeMicrosoftToFake` in `support/ui.ts`). Mattermost's `SiteURL` stays `http://localhost:8065`.
- The fake exposes a control API under `/__control/` (reset, config, token expiry and revocation, subscriptions and notifications, requests log, events per user). `support/fake.ts` wraps it.
- `fake-microsoft/contract.ts` lists the fields each response must contain. The fake validates its own responses and the plugin's requests against it. Violations are listed at `/__control/violations`, and global teardown fails the run if there are any.

Test users are `alice@contoso.test` (connected to Mattermost) and `bob@contoso.test` (invited to events).

## Running

From the repo root:

```sh
make e2e-env-up   # Postgres, the fake and Mattermost on http://localhost:8065
make e2e    # build the plugin bundle, install deps, run the suite
make e2e-env-down    # stop and wipe the server
```

Or from this directory, after `make dist` and `make e2e-env-up`:

```sh
npm ci && npx playwright install chromium
npm test              # headless
npm run test:headed
npm run test:ui
```

Fixture events live in the fake and are removed by each spec. `POST /__control/reset` clears everything, and global setup calls it.

### Coverage

| Spec | Covers |
| --- | --- |
| `connect` | Connect and disconnect through the login popup, forced consent |
| `behavior` | Wrong password, denied consent, access-token refresh, revoked refresh token |
| `commands` | `help`, `info`, `connect`, `today`, `tomorrow`, `viewcal`, `settings`, `showcals` |
| `events` | Create-event modal, invitation reaching the attendee, form validation |
| `availability` | Status sync (`avail`) sets Do Not Disturb from a running meeting |
| `sidebar` | Channel header button and the experimental calendar sidebar |
| `webhooks` | Subscription handshake and `clientState`, new and updated event DMs, forged notification rejected, renewal |

## Contract check against the real service

`npm run contract:compare` runs the same Graph scenarios (token, create, get and list events, mailbox settings, calendars, batch `getSchedule`, error shapes, delete) against the fake and a live tenant. It fails when a response is missing a contract field or when the fake and the live service answer with different status codes. The nightly workflow `e2e-contract.yml` runs it.

```sh
npm run fake              # start the fake on http://127.0.0.1:8080
npm run contract:fake     # scenarios against the fake only
npm run contract:live     # scenarios against the live tenant only
npm run contract:compare  # both, and report drift
```

When Microsoft changes a response, the nightly job goes red. Update the fake and `fake-microsoft/contract.ts` together, then rerun the suite.

### Live tenant setup (one-time, by a tenant admin)

Use a **dedicated** tenant, never the corporate one.

1. Create two licensed users (Exchange Online) so Graph does not return `MailboxNotEnabledForRESTAPI`.
2. Register an app with the application permissions `Calendars.ReadWrite` and `MailboxSettings.Read`, and grant admin consent.
3. Store these in the GitHub environment `e2e-entra`, or in an untracked `e2e/playwright/.env` for local runs:

| Variable | Value |
| --- | --- |
| `MSCAL_E2E_TENANT_ID` | Directory (tenant) ID |
| `MSCAL_E2E_CLIENT_ID` | App client ID |
| `MSCAL_E2E_CLIENT_SECRET` | App client secret |
| `MSCAL_E2E_USER1_EMAIL` / `MSCAL_E2E_USER2_EMAIL` | The two test mailboxes |

Rotate the secret by creating a new one on the app and updating `MSCAL_E2E_CLIENT_SECRET`.

## Notes

- Do not run the suite against a Mattermost server you care about. Setup installs the plugin, rewrites its settings and, unless the server is `localhost`, refuses to delete other teams.
- The plugin refreshes access tokens that expire within five minutes, so the token specs use a 330 second token and wait about 35 seconds.
- `make dist` can leave `plugin.json` owner-only inside the bundle. `installPlugin` repacks the bundle with readable permissions before uploading.
