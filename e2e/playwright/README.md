# Microsoft Calendar plugin E2E tests

Playwright suite that drives the plugin in a real browser against a disposable Mattermost server and a real Microsoft Entra ID tenant. Unlike the GitLab plugin suite, the identity provider cannot run in a container, so the suite signs in as dedicated test users in a dedicated tenant and reads and writes their real calendars through Microsoft Graph.

Built on [`@mattermost/playwright-lib`](https://www.npmjs.com/package/@mattermost/playwright-lib), so no `mattermost` monorepo checkout is needed.

## How it works

One pre-registered Entra app (the test app) supplies the OAuth client ID and secret. Its only redirect URI is `http://localhost:8065/plugins/com.mattermost.mscalendar/oauth2/complete`, and Mattermost's `SiteURL` is `http://localhost:8065` everywhere (local, CI, Cursor Cloud Agents). The suite never modifies the Entra app.

Setup writes the app credentials into the plugin settings. Specs sign in to Mattermost as a regular member, connect that member to test user 1 through the real Microsoft login page, and create fixture events with an app-only Graph token. Every fixture event subject starts with `[e2e-<runId>]`, and teardown deletes only the events of its own run.

### Out of scope: Graph change notifications

Microsoft Graph only creates a subscription when it can reach the notification URL, and GitHub-hosted runners cannot receive inbound traffic. Subscriptions, event DMs and renewals are therefore not covered here; the Go unit tests in `msgraph/handle_webhook_test.go` and `calendar/engine/notification_test.go` cover them. Specs never enable event subscriptions.

## Prerequisites (one-time, by a tenant admin)

Use a **dedicated** tenant, never the corporate one.

1. **Tenant:** disable Security defaults (Entra admin center > Identity > Overview > Properties > Manage security defaults), so test users can sign in with a password only. Make sure no Conditional Access policy requires MFA for the test users or the app.
2. **Test users:** create two users, each with an Exchange Online license (for example Microsoft 365 Business Basic). Without a license Graph returns `MailboxNotEnabledForRESTAPI`. User 2 is only invited to events.
3. **Test app:** register an app (single tenant) with:
   - Redirect URI of type **Web**: `http://localhost:8065/plugins/com.mattermost.mscalendar/oauth2/complete`
   - A client secret (note its expiry date)
   - Delegated permissions: `offline_access`, `User.Read`, `Calendars.ReadWrite`, `Calendars.ReadWrite.Shared`, `MailboxSettings.Read`
   - Application permissions: `Calendars.Read`, `MailboxSettings.Read` (status sync and the daily summary) and `Calendars.ReadWrite` (fixture events)
   - Admin consent granted for all of the above
4. **Optional hardening:** limit the app's application permissions to the two test mailboxes with an Exchange Online RBAC for Applications management scope.
5. **Secrets:** the GitHub environment `e2e-entra` and, for Cursor Cloud Agents, an environment-scoped Runtime Secret set, both holding:

| Variable | Value |
| --- | --- |
| `MSCAL_E2E_TENANT_ID` | Directory (tenant) ID |
| `MSCAL_E2E_CLIENT_ID` | Test app client ID |
| `MSCAL_E2E_CLIENT_SECRET` | Test app client secret |
| `MSCAL_E2E_USER1_EMAIL` / `MSCAL_E2E_USER1_PASSWORD` | Test user 1 (connected to Mattermost) |
| `MSCAL_E2E_USER2_EMAIL` / `MSCAL_E2E_USER2_PASSWORD` | Test user 2 (invited to events) |

For local runs, put the same variables in an untracked `e2e/playwright/.env`.

### Rotating the client secret

Create a new secret on the test app, update `MSCAL_E2E_CLIENT_SECRET` in the GitHub environment and in Cursor, then delete the old secret.

## Running

From the repo root:

```sh
make e2e-start   # Postgres and Mattermost on http://localhost:8065
make e2e-test    # build the plugin bundle, install deps, run the suite
make e2e-stop    # stop and wipe the server
```

Or from this directory, after `make dist` and `make e2e-start`:

```sh
npm ci && npx playwright install chromium
npm test              # headless
npm run test:headed   # watch the browser
npm run test:ui       # Playwright UI mode
npm run show-report   # open the last HTML report
npm run cleanup       # delete leftover [e2e-*] events older than 24 hours
```

Port 8065 must be free, because the Microsoft redirect URI is fixed to `http://localhost:8065`. Setup deletes every other team the sysadmin belongs to, but only on `http://localhost:8065` unless `PW_DELETE_OTHER_TEAMS=true` is set.

### Parallel runs

All runs share the same two mailboxes. CI serializes runs with a `concurrency` group. Local and Cloud Agent runs should not overlap: the availability spec reads the user's status, which another run's "happening now" event can change.

### Coverage

| Spec | Covers |
| --- | --- |
| `connect` | OAuth connect and disconnect, including the forced-consent screen |
| `commands` | `help`, `info`, `settings`, `today`, `tomorrow`, `summary`, `viewcal`, `showcals`, usage errors |
| `events` | Create an event from the modal and `/mscalendar event create`, invite user 2 |
| `availability` | `/mscalendar avail` sets the status from a running event |
| `sidebar` | Experimental calendar sidebar and channel header button |

### Cursor Cloud Agents

`.cursor/environment.json` builds an Ubuntu image with Docker, Go, Node and Playwright, forwards port 8065 and starts the Compose stack. Add the variables above as Runtime Secrets scoped to the environment. If your team restricts egress, allow `login.microsoftonline.com`, `login.live.com`, `aadcdn.msauth.net`, `aadcdn.msftauth.net`, `graph.microsoft.com` and Docker Hub.
