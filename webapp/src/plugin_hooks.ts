// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {openCreateEventModal} from './actions';
import {isUserConnected} from './selectors';

// import {openCreateModalWithoutPost, openChannelSettings, sendEphemeralPost, openDisconnectModal, handleConnectFlow, getConnected} from '../actions';
// import {isUserConnected, getInstalledInstances, getPluginSettings, getUserConnectedInstances, instanceIsInstalled} from '../selectors';

type ContextArgs = {channel_id: string};

const createEventCommand = '/z';

// const createEventCommand = '/gcal createevent';

interface Store {
    dispatch(action: {type: string}): void;
    getState(): object;
}

export default class Hooks {
    private store: Store;

    constructor(store: Store) {
        this.store = store;
    }

    slashCommandWillBePostedHook = (rawMessage: string, contextArgs: ContextArgs) => {
        let message;
        if (rawMessage) {
            message = rawMessage.trim();
        }

        if (!message) {
            return Promise.resolve({message, args: contextArgs});
        }

        if (message.startsWith(createEventCommand)) {
            return this.handleCreateEventSlashCommand(message, contextArgs);
        }

        return Promise.resolve({message, args: contextArgs});
    };

    handleCreateEventSlashCommand = async (message: string, contextArgs: ContextArgs) => {
        if (!(await this.checkUserIsConnected())) {
            return Promise.resolve({});
        }

        this.store.dispatch(openCreateEventModal(contextArgs.channel_id));
        return Promise.resolve({});
    };

    checkUserIsConnected = async (): Promise<boolean> => {
        return true;

        if (!isUserConnected(this.store.getState())) {
            await this.store.dispatch(getConnected());
            if (!isUserConnected(this.store.getState())) {
                this.store.dispatch(sendEphemeralPost('Your Mattermost account is not connected to Jira. Please use `/jira connect` to connect your account, then try again.'));
                return false;
            }
        }

        return true;
    };
}
