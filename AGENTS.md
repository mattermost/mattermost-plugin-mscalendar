# Agent instructions

## Cursor Cloud specific instructions

Cloud agent VMs have no Docker. Mattermost 12 runs natively from `/opt/mattermost` against a local Postgres, started by the environment start script. `e2e/playwright/scripts/native-env-up.sh` runs after it and brings up everything else the E2E suite needs, so a fresh VM is normally ready at boot:

- The fake Microsoft Entra ID and Graph server on `:8080` (HTTP) and `:443` (HTTPS, tmux session `e2e-fake-microsoft`).
- `/etc/hosts` aliases for `login.microsoftonline.com` and `graph.microsoft.com`, a throwaway CA in `/certs`, and Mattermost (tmux session `mattermost`) restarted with `SSL_CERT_FILE` pointing at it.
- The plugin settings in `/opt/mattermost/config/config.json` pointing at the fake tenant, with the camel-case settings the start script seeds removed. They would otherwise shadow the suite's lowercase keys and fail every sign-in with `AADSTS700016`.
- `e2e/playwright/.env` (untracked) with `PW_ADMIN_PASSWORD`.

If something is down, run `make e2e-env-up-native`. It is safe to run repeatedly.

| Task | Command |
| --- | --- |
| Lint | `make check-style` |
| Go and webapp unit tests | `make test` |
| Build the bundle | `make dist` (generates `server/manifest.go`, so a bare `go build ./...` fails before the first `make apply`) |
| Deploy to the local server | `MM_LOCALSOCKETPATH=/var/tmp/mattermost_local.socket make deploy` |
| Full E2E suite | `make e2e` |
| One E2E spec | `cd e2e/playwright && npx playwright test tests/commands.spec.ts` |

Notes:

- The suite installs the freshest bundle in `dist/` itself (`make e2e` builds it first), and runs against `http://localhost:8065`, which it treats as disposable. It archives every team except its own, so do not keep data you care about on this server.
- Admin login is `sysadmin` with the password in `/opt/mattermost/.admin-password`.
- Mattermost logs are in `/tmp/mattermost.log` and the fake's in `/tmp/fake-microsoft.log`. Failed specs leave traces, screenshots and videos in `e2e/playwright/results/output`.
