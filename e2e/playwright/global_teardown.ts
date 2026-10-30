// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './support/env';

import {fake} from './support/fake';
import {disconnectRegularUser} from './support/plugin';

export default async function globalTeardown() {
    await disconnectRegularUser();

    const violations = await fake.violations();
    if (violations.length > 0) {
        throw new Error(`The fake saw requests or produced responses that break the Graph contract:\n${violations.join('\n')}`);
    }
}
