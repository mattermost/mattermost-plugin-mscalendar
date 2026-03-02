import {GlobalState} from '@mattermost/types/store';

import {RemoteEvent} from '@/types/calendar';

import manifest from '../manifest';

interface PluginState {
    events: {
        items: RemoteEvent[];
        loading: boolean;
        error: string | null;
    };
}

function getPluginState(state: GlobalState): PluginState {
    return (state as any)[`plugins-${manifest.id}`] || {events: {items: [], loading: false, error: null}};
}

export function getCalendarEvents(state: GlobalState): RemoteEvent[] {
    return getPluginState(state).events.items;
}

export function getCalendarEventsLoading(state: GlobalState): boolean {
    return getPluginState(state).events.loading;
}

export function getCalendarEventsError(state: GlobalState): string | null {
    return getPluginState(state).events.error;
}
