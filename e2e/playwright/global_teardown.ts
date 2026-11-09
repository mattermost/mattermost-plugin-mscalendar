// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './support/env';

import {loginNonAdminUser} from './support/admin';
import {fake} from './support/fake';
import {disconnectMember} from './support/plugin';

export default async function globalTeardown() {
    await disconnectMember();
    await disconnectMember(loginNonAdminUser);

    const violations = await fake.violations();
    if (violations.length > 0) {
        throw new Error(`The fake saw requests or produced responses that break the Graph contract:\n${violations.join('\n')}`);
    }
}
