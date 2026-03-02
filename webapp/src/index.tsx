import React, {useEffect} from 'react';

import {Store, Action} from 'redux';

import {GlobalState} from '@mattermost/types/store';

import {PluginRegistry} from '@/types/mattermost-webapp';

import CalendarSidebar from './components/calendar_sidebar';
import ChannelHeaderIcon from './components/channel_header_icon/channel_header_icon';
import reducer from './reducers';
import client from './client/client';
import {PluginId} from './plugin_id';

export default class Plugin {
    private haveSetupUI = false;

    private finishedSetupUI = () => {
        this.haveSetupUI = true;
    };

    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        this.haveSetupUI = false;

        registry.registerReducer(reducer);

        const setup = async () => {
            let providerConfig: any = {};
            try {
                providerConfig = await client.getProviderConfiguration();
            } catch {
                // If fetch fails, default to UI disabled
            }

            if (providerConfig?.Features?.EnableExperimentalUI) {
                const {showRHSPlugin} = registry.registerRightHandSidebarComponent(
                    CalendarSidebar,
                    'Calendar',
                );

                registry.registerChannelHeaderButtonAction(
                    <ChannelHeaderIcon/>,
                    () => store.dispatch(showRHSPlugin),
                    'Calendar',
                    'Toggle calendar sidebar',
                );
            }
        };

        registry.registerRootComponent(() => (
            <SetupUI
                setup={setup}
                haveSetupUI={this.haveSetupUI}
                finishedSetupUI={this.finishedSetupUI}
            />
        ));
    }
}

interface SetupUIProps {
    setup: () => Promise<void>;
    haveSetupUI: boolean;
    finishedSetupUI: () => void;
}

const SetupUI = ({setup, haveSetupUI, finishedSetupUI}: SetupUIProps) => {
    useEffect(() => {
        if (!haveSetupUI) {
            setup().then(() => {
                finishedSetupUI();
            });
        }
    }, []);

    return null;
};

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
    }
}

window.registerPlugin(PluginId, new Plugin());
