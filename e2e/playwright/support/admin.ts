// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './env';

import {Client4} from '@mattermost/client';
import type {UserProfile} from '@mattermost/types/users';
import {testConfig} from '@mattermost/playwright-lib';

import {TEAM_NAME} from './constants';

const DISPOSABLE_SERVER_URL = 'http://localhost:8065';

export function isDisposableServer() {
    return new URL(testConfig.baseURL).origin === DISPOSABLE_SERVER_URL;
}

type MemberAccount = {username: string; email: string};

// A fixed user, rather than a random one per test, keeps a single Microsoft connection across specs.
// Setup lists this member as a plugin admin.
const REGULAR_USER: MemberAccount = {
    username: 'mscalendar-e2e-user',
    email: 'mscalendar-e2e-user@example.com',
};

const NON_ADMIN_USER: MemberAccount = {
    username: 'mscalendar-e2e-non-admin',
    email: 'mscalendar-e2e-non-admin@example.com',
};

// The default password is public, so other servers must supply their own.
function regularUserPassword() {
    const password = process.env.PW_REGULAR_USER_PASSWORD;
    if (password) {
        return password;
    }
    if (!isDisposableServer()) {
        throw new Error(`Set PW_REGULAR_USER_PASSWORD to run against ${testConfig.baseURL}.`);
    }
    return 'Ms-E2e-Member#2026';
}

async function login(username: string, password: string) {
    const client = new Client4();
    client.setUrl(testConfig.baseURL);
    const profile = await client.login(username, password);
    return {client, user: {...profile, password}};
}

// loginAdmin works outside test fixtures (e.g. in global setup), unlike pw.getAdminClient().
export function loginAdmin() {
    return login(testConfig.adminUsername, testConfig.adminPassword);
}

export type MemberLogin = () => ReturnType<typeof login>;

export function loginRegularUser() {
    return login(REGULAR_USER.username, regularUserPassword());
}

export function loginNonAdminUser() {
    return login(NON_ADMIN_USER.username, regularUserPassword());
}

export function ensureRegularUser(adminClient: Client4) {
    return ensureMember(adminClient, REGULAR_USER);
}

export function ensureNonAdminUser(adminClient: Client4) {
    return ensureMember(adminClient, NON_ADMIN_USER);
}

async function ensureMember(adminClient: Client4, account: MemberAccount): Promise<UserProfile> {
    const password = regularUserPassword();
    let user: UserProfile;
    try {
        user = await adminClient.getUserByUsername(account.username);
    } catch {
        user = await adminClient.createUser({...account, password} as UserProfile, '', '');
    }
    if (user.roles.split(' ').includes('system_admin')) {
        throw new Error(`${account.username} must not be a system admin`);
    }

    const team = await adminClient.getTeamByName(TEAM_NAME);
    await adminClient.addToTeam(team.id, user.id);
    return user;
}
