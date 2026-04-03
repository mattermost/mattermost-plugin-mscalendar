import React, {useEffect, useRef} from 'react';

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

        const thunkDispatch = store.dispatch as AppDispatch;

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

        const setup = async () => {
            await thunkDispatch(getProviderConfiguration());

            const providerConfig = getProviderConfigSelector(store.getState());
            if (!providerConfig) {
                throw new Error('Failed to load provider configuration');
            }

            if (providerConfig.Features?.EnableExperimentalUI) {
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
    const startedRef = useRef(false);

    useEffect(() => {
        if (!haveSetupUI && !startedRef.current) {
            startedRef.current = true;
            setup().
                then(() => {
                    finishedSetupUI();
                }).
                catch((error) => {
                    startedRef.current = false;

                    // eslint-disable-next-line no-console
                    console.error('Plugin setup failed:', error);
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
