import React, {useEffect} from 'react';

import {Store, Action} from 'redux';

import {GlobalState} from '@mattermost/types/lib/store';

import {PluginRegistry} from '@/types/mattermost-webapp';

import {PluginId} from './plugin_id';

import Hooks from './plugin_hooks';
import reducer from './reducers';

import CreateEventModal from './components/modals/create_event_modal';

export default class Plugin {
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        registry.registerReducer(reducer);

        const hooks = new Hooks(store);
        registry.registerSlashCommandWillBePostedHook(hooks.slashCommandWillBePostedHook);

        const setup = () => {
            registry.registerRootComponent(CreateEventModal);
        };

        registry.registerRootComponent(() => <SetupUI setup={setup}/>);

        // reminder to set up site url for any API calls
        // and i18n
    }
}

const SetupUI = ({setup}) => {
    useEffect(() => {
        setup();
    }, []);

    return null;
};

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
    }
}

window.registerPlugin(PluginId, new Plugin());
