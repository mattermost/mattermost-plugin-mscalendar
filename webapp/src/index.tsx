import React, {useEffect} from 'react';

import {Action, Store} from 'redux';

import {GlobalState} from '@mattermost/types/store';

import type {AppDispatch} from '@/hooks';

import {PluginRegistry} from '@/types/mattermost-webapp';

import {PluginId} from './plugin_id';

import Hooks from './plugin_hooks';
import reducer from './reducers';

import CalendarSidebar from './components/calendar_sidebar';
import ChannelHeaderIcon from './components/channel_header_icon/channel_header_icon';
import CreateEventModal from './components/modals/create_event_modal';
import {getProviderConfiguration, handleConnect, handleDisconnect, openCreateEventModal} from './actions';
import {getProviderConfiguration as getProviderConfigSelector} from './selectors';

export default class Plugin {
    private haveSetupUI = false;

    private finishedSetupUI = () => {
        this.haveSetupUI = true;
    };

    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        this.haveSetupUI = false;

        registry.registerReducer(reducer);

        const hooks = new Hooks(store);
        registry.registerSlashCommandWillBePostedHook(hooks.slashCommandWillBePostedHook);

        const setup = async () => {
            const thunkDispatch = store.dispatch as AppDispatch;
            await thunkDispatch(getProviderConfiguration());

            const providerConfig = getProviderConfigSelector(store.getState());

            if (providerConfig?.Features?.EnableExperimentalUI) {
                const {toggleRHSPlugin} = registry.registerRightHandSidebarComponent(
                    CalendarSidebar,
                    'Calendar',
                );

                registry.registerChannelHeaderButtonAction(
                    <ChannelHeaderIcon/>,
                    () => store.dispatch(toggleRHSPlugin),
                    'Calendar',
                    'Toggle calendar sidebar',
                );
            }

            registry.registerChannelHeaderMenuAction(
                <span>{'Create calendar event'}</span>,
                async (channelID) => {
                    if (await hooks.checkUserIsConnected()) {
                        thunkDispatch(openCreateEventModal(channelID));
                    }
                },
            );

            registry.registerRootComponent(CreateEventModal);

            registry.registerWebSocketEventHandler(`custom_${PluginId}_connected`, handleConnect(store));
            registry.registerWebSocketEventHandler(`custom_${PluginId}_disconnected`, handleDisconnect(store));
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
            setup().
                catch((error) => {
                    // eslint-disable-next-line no-console
                    console.error('Plugin setup failed:', error);
                }).
                finally(() => {
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
