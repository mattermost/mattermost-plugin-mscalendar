// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import './support/env';

import {baseGlobalSetup} from '@mattermost/playwright-lib';

import {startRun} from './support/credentials';
import {prepareServer} from './support/plugin';

export default async function globalSetup() {
    await baseGlobalSetup();
    const runId = startRun();
    // eslint-disable-next-line no-console
    console.log(`E2E run id: ${runId}`);
    await prepareServer();
}
