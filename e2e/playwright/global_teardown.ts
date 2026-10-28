// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './support/env';

import {currentRunId, loadCredentials} from './support/credentials';
import {deleteEventsWithPrefix} from './support/graph';
import {disconnectRegularUser} from './support/plugin';

export default async function globalTeardown() {
    const runId = currentRunId();
    if (!runId) {
        return;
    }

    const {user1, user2} = loadCredentials();
    await Promise.all([user1, user2].map((user) => deleteEventsWithPrefix(user.email, `[e2e-${runId}]`)));
    await disconnectRegularUser();
}
