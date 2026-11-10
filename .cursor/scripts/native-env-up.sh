#!/usr/bin/env bash
# Docker-free equivalent of `make e2e-env-up` for Cursor cloud agent VMs, which have no Docker and
# run Mattermost natively from /opt/mattermost against a local Postgres.
#
# It is idempotent: it brings up the fake Microsoft server on :8080/:443, makes Mattermost
# trust it, points the plugin at it, and leaves e2e/playwright/.env with the admin password,
# so `make e2e` works with no further setup.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(cd "${SCRIPT_DIR}/../../e2e/playwright" && pwd)"
MM_DIR=/opt/mattermost
MM_CONFIG="${MM_DIR}/config/config.json"
MM_ADMIN_PASSWORD_FILE="${MM_DIR}/.admin-password"
CERT_DIR=/certs
CA_BUNDLE="${CERT_DIR}/ca-bundle.crt"
MM_LAUNCHER=/tmp/start-mm-e2e.sh
MM_URL=http://127.0.0.1:8065
FAKE_URL=http://127.0.0.1:8080
MICROSOFT_HOSTS="login.microsoftonline.com graph.microsoft.com"

TMUX=(tmux)
[[ -f /exec-daemon/tmux.portal.conf ]] && TMUX=(tmux -f /exec-daemon/tmux.portal.conf)

log() { echo "[native-e2e] $*"; }

mm_healthy() { curl -sf "${MM_URL}/api/v4/system/ping" >/dev/null 2>&1; }
fake_healthy() { curl -sf "${FAKE_URL}/__control/health" >/dev/null 2>&1; }
mm_pid() { pgrep -x mattermost | head -1 || true; }

wait_for() {
    local description=$1 attempts=$2
    shift 2
    for _ in $(seq 1 "${attempts}"); do
        if "$@"; then
            return 0
        fi
        sleep 2
    done
    echo "Timed out waiting for ${description}" >&2
    return 1
}

[[ -x "${MM_DIR}/bin/mattermost" ]] || { echo "${MM_DIR}/bin/mattermost not found; this script needs the native Mattermost install." >&2; exit 1; }

if ! mm_healthy; then
    if [[ -f /opt/cursor/cloud-agent-start.sh ]]; then
        log "Mattermost is down; running the environment start script"
        bash /opt/cursor/cloud-agent-start.sh
    else
        echo "Mattermost is not running on ${MM_URL}" >&2
        exit 1
    fi
fi

# The fake and Mattermost both need the throwaway CA and the Microsoft host aliases.
sudo mkdir -p "${CERT_DIR}"
sudo chown "$(id -u):$(id -g)" "${CERT_DIR}"
sh "${E2E_DIR}/fake-microsoft/gen-certs.sh" >/tmp/gen-certs.log 2>&1 || { cat /tmp/gen-certs.log >&2; exit 1; }

if ! grep -q "login.microsoftonline.com" /etc/hosts; then
    echo "127.0.0.1 ${MICROSOFT_HOSTS}" | sudo tee -a /etc/hosts >/dev/null
fi
sudo sysctl -q -w net.ipv4.ip_unprivileged_port_start=443

if ! fake_healthy; then
    log "Starting the fake Microsoft server"
    "${TMUX[@]}" kill-session -t e2e-fake-microsoft 2>/dev/null || true
    "${TMUX[@]}" new-session -d -s e2e-fake-microsoft -c "${E2E_DIR}" \
        "FAKE_CERT_DIR=${CERT_DIR} FAKE_MATTERMOST_URL=http://localhost:8065 node fake-microsoft/server.ts 2>&1 | tee /tmp/fake-microsoft.log"
    wait_for "the fake Microsoft server" 15 fake_healthy
fi

# The environment start script seeds camel-case plugin settings (a made-up client id) that shadow the
# lowercase keys the suite writes, so the plugin would sign in against the wrong app. Keep only the fake's.
config_is_current() {
    python3 - "${MM_CONFIG}" <<'PY'
import json, sys
settings = json.load(open(sys.argv[1]))["PluginSettings"].get("Plugins", {}).get("com.mattermost.mscalendar", {})
stale = {"EnableExperimentalUI", "OAuth2Authority", "OAuth2ClientId", "OAuth2ClientSecret", "OAuth2ForceConsent"}
fake = {"oauth2authority": "fake-tenant", "oauth2clientid": "fake-client-id", "oauth2clientsecret": "fake-client-secret"}
sys.exit(0 if not stale & settings.keys() and all(settings.get(k) == v for k, v in fake.items()) else 1)
PY
}

mm_trusts_fake() {
    local pid
    pid=$(mm_pid)
    [[ -n "${pid}" ]] && tr '\0' '\n' <"/proc/${pid}/environ" 2>/dev/null | grep -qx "SSL_CERT_FILE=${CA_BUNDLE}"
}

if ! mm_trusts_fake || ! config_is_current; then
    log "Restarting Mattermost with the fake CA and plugin settings"
    "${TMUX[@]}" kill-session -t mattermost 2>/dev/null || true
    pid=$(mm_pid)
    if [[ -n "${pid}" ]]; then
        kill "${pid}"
        for _ in $(seq 1 30); do
            kill -0 "${pid}" 2>/dev/null || break
            sleep 1
        done
    fi
    rm -f /var/tmp/mattermost_local.socket

    python3 - "${MM_CONFIG}" <<'PY'
import json, sys
path = sys.argv[1]
config = json.load(open(path))
plugins = config["PluginSettings"].setdefault("Plugins", {})
settings = plugins.setdefault("com.mattermost.mscalendar", {})
for key in ("EnableExperimentalUI", "OAuth2Authority", "OAuth2ClientId", "OAuth2ClientSecret", "OAuth2ForceConsent"):
    settings.pop(key, None)
settings.update({
    "oauth2authority": "fake-tenant",
    "oauth2clientid": "fake-client-id",
    "oauth2clientsecret": "fake-client-secret",
    "oauth2forceconsent": False,
    "enableexperimentalui": True,
})
with open(path, "w") as fh:
    json.dump(config, fh, indent=4)
    fh.write("\n")
PY

    # The same settings as e2e/playwright/docker-compose.yml, so the UI is not covered by onboarding overlays.
    cat >"${MM_LAUNCHER}" <<EOF
set -a
source ${MM_DIR}/cloud-agent.env
set +a
export SSL_CERT_FILE=${CA_BUNDLE}
export MM_SERVICESETTINGS_ENABLEONBOARDINGFLOW=false
export MM_SERVICESETTINGS_ENABLETUTORIAL=false
export MM_SERVICESETTINGS_ENABLESECURITYFIXALERT=false
export MM_PLUGINSETTINGS_AUTOMATICPREPACKAGEDPLUGINS=false
export MM_RATELIMITSETTINGS_ENABLE=false
export MM_ANNOUNCEMENTSETTINGS_ADMINNOTICESENABLED=false
export MM_ANNOUNCEMENTSETTINGS_USERNOTICESENABLED=false
cd ${MM_DIR}
exec ./bin/mattermost server >> /tmp/mattermost.log 2>&1
EOF
    "${TMUX[@]}" new-session -d -s mattermost "bash ${MM_LAUNCHER}"
    wait_for "Mattermost" 60 mm_healthy
    wait_for "the Mattermost local socket" 30 test -S /var/tmp/mattermost_local.socket
fi

# Untracked file read by support/env.ts so `make e2e` needs no exported variables.
ENV_FILE="${E2E_DIR}/.env"
ADMIN_PASSWORD="$(tr -d '\n' <"${MM_ADMIN_PASSWORD_FILE}")"
if ! grep -qx "PW_ADMIN_PASSWORD=${ADMIN_PASSWORD}" "${ENV_FILE}" 2>/dev/null; then
    ( umask 077; grep -v '^PW_ADMIN_PASSWORD=' "${ENV_FILE}" 2>/dev/null >"${ENV_FILE}.tmp" || true
      echo "PW_ADMIN_PASSWORD=${ADMIN_PASSWORD}" >>"${ENV_FILE}.tmp"
      mv "${ENV_FILE}.tmp" "${ENV_FILE}" )
fi

log "Ready: Mattermost ${MM_URL}, fake Microsoft ${FAKE_URL}. Run \`make e2e\` from the repo root."
