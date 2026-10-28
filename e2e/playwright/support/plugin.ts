// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import fs from 'node:fs';
import path from 'node:path';

import './env';

import type {Client4} from '@mattermost/client';
import {testConfig} from '@mattermost/playwright-lib';

import {ensureRegularUser, isDisposableServer, loginAdmin, loginRegularUser} from './admin';
import {COMMAND, PLUGIN_ID, TEAM_NAME, TOWN_SQUARE} from './constants';
import {loadCredentials} from './credentials';
import {deleteEventsWithPrefix} from './graph';

const DIST_DIR = path.resolve(__dirname, '../../../dist');
type PluginSettingValues = Record<string, unknown>;

export function findPluginBundle(): string {
    const bundles = fs.existsSync(DIST_DIR) ? fs.readdirSync(DIST_DIR).filter((f) => f.startsWith(PLUGIN_ID) && f.endsWith('.tar.gz')) : [];
    if (bundles.length === 0) {
        throw new Error(`No plugin bundle found in ${DIST_DIR}. Run \`make dist\` from the repo root first.`);
    }

    return bundles.
        map((f) => path.join(DIST_DIR, f)).
        sort((a, b) => fs.statSync(b).mtimeMs - fs.statSync(a).mtimeMs)[0];
}

export async function installPlugin(client: Client4, bundlePath: string) {
    const bytes = fs.readFileSync(bundlePath);
    const file = new File([bytes], path.basename(bundlePath), {type: 'application/gzip'});
    await client.uploadPlugin(file, true);
}

export async function restartPlugin(client: Client4) {
    await client.disablePlugin(PLUGIN_ID);
    await client.enablePlugin(PLUGIN_ID);
    await waitForPlugin(client);
}

async function waitForPlugin(client: Client4) {
    const deadline = Date.now() + 20000;
    while (Date.now() < deadline) {
        const plugins = await client.getPlugins();
        if (plugins.active.some((plugin) => plugin.id === PLUGIN_ID)) {
            return;
        }
        // eslint-disable-next-line no-await-in-loop
        await new Promise((resolve) => setTimeout(resolve, 500));
    }
    throw new Error(`Timed out waiting for ${PLUGIN_ID} to become active`);
}

export async function patchPluginSettings(client: Client4, patch: PluginSettingValues) {
    const config = await client.getConfig();
    const plugins = (config.PluginSettings.Plugins ?? {}) as Record<string, PluginSettingValues>;
    plugins[PLUGIN_ID] = {...(plugins[PLUGIN_ID] ?? {}), ...patch};
    config.PluginSettings.Plugins = plugins as typeof config.PluginSettings.Plugins;
    await client.updateConfig(config);
}

// The plugin restarts with the new settings, so callers that need it running
// afterwards should use restartPlugin or wait.
export async function configurePlugin(client: Client4, adminUserId: string, forceConsent = false) {
    const credentials = loadCredentials();
    await patchPluginSettings(client, {
        oauth2authority: credentials.tenantId,
        oauth2clientid: credentials.clientId,
        oauth2clientsecret: credentials.clientSecret,
        oauth2forceconsent: forceConsent,
        enableexperimentalui: true,

        // The admin-only commands (avail, showcals) are run by the connected regular member.
        adminuserids: adminUserId,
    });
}

// ensureTeam keeps a single team so the channel-header buttons render predictably.
export async function ensureTeam(client: Client4) {
    const teams = await client.getMyTeams();
    if (!teams.some((team) => team.name === TEAM_NAME)) {
        await client.createTeam({
            name: TEAM_NAME,
            display_name: 'MS Calendar E2E',
            type: 'O',
        } as Parameters<Client4['createTeam']>[0]);
    }

    const others = (await client.getMyTeams()).filter((team) => team.name !== TEAM_NAME);
    if (others.length === 0) {
        return;
    }

    if (!isDisposableServer() && process.env.PW_DELETE_OTHER_TEAMS !== 'true') {
        throw new Error(`Refusing to delete teams ${others.map((team) => team.name).join(', ')} on ${testConfig.baseURL}. ` +
            'Leave them first, or set PW_DELETE_OTHER_TEAMS=true to allow deleting them on this non-disposable server.');
    }
    await Promise.all(others.map((team) => client.deleteTeam(team.id)));
}

export async function executeCommand(client: Client4, command: string) {
    const team = await client.getTeamByName(TEAM_NAME);
    const channel = await client.getChannelByName(team.id, TOWN_SQUARE);
    return client.executeCommand(command, {channel_id: channel.id, team_id: team.id});
}

export async function disconnectRegularUser() {
    try {
        const {client} = await loginRegularUser();
        await executeCommand(client, `${COMMAND} disconnect`);
    } catch {
        // The user may not exist yet or may already be disconnected.
    }
}

const STALE_EVENT_AGE_MS = 24 * 60 * 60 * 1000;

// Removes fixtures that cancelled runs left behind, keeping anything a concurrent run just created.
export async function sweepStaleEvents() {
    const {user1, user2} = loadCredentials();
    await Promise.all([user1, user2].map((user) => deleteEventsWithPrefix(user.email, '[e2e-', STALE_EVENT_AGE_MS)));
}

export async function prepareServer() {
    const {client} = await loginAdmin();
    await ensureTeam(client);
    const regularUser = await ensureRegularUser(client);

    const bundle = findPluginBundle();
    // eslint-disable-next-line no-console
    console.log(`Installing ${bundle}`);
    await installPlugin(client, bundle);
    await configurePlugin(client, regularUser.id);
    await restartPlugin(client);

    // A previous run may have left the member connected to Microsoft.
    await disconnectRegularUser();
    await sweepStaleEvents();
}
