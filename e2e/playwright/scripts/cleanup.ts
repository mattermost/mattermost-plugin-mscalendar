// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// Deletes leftover [e2e-*] events from both test mailboxes.
//   npm run cleanup                      events older than 24 hours
//   npm run cleanup -- --run-id <id>     every event of one run (CI's always-run step)
import '../support/env';

import {loadCredentials} from '../support/credentials';
import {deleteEventsWithPrefix} from '../support/graph';

const DAY_MS = 24 * 60 * 60 * 1000;

async function main() {
    const flag = process.argv.indexOf('--run-id');
    const runId = flag === -1 ? '' : process.argv[flag + 1];
    const prefix = runId ? `[e2e-${runId}]` : '[e2e-';
    const olderThanMs = runId ? 0 : DAY_MS;

    const {user1, user2} = loadCredentials();
    for (const user of [user1, user2]) {
        // eslint-disable-next-line no-await-in-loop
        const count = await deleteEventsWithPrefix(user.email, prefix, olderThanMs);
        // eslint-disable-next-line no-console
        console.log(`Deleted ${count} event(s) matching ${prefix} from ${user.email}`);
    }
}

main().catch((err) => {
    // eslint-disable-next-line no-console
    console.error(err);
    process.exit(1);
});
