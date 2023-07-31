import React, {useEffect} from 'react';

import {Store, Action} from 'redux';

import {GlobalState} from '@mattermost/types/lib/store';

import manifest from '../../plugin.json';

import {PluginRegistry} from '@/types/mattermost-webapp';
import CreateIssueModal from './components/modals/create_event_modal';
import Hooks from './plugin_hooks';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/

        const hooks = new Hooks(store);
        registry.registerSlashCommandWillBePostedHook(hooks.slashCommandWillBePostedHook);

        const setup = () => {
            registry.registerRootComponent(CreateIssueModal);
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
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
