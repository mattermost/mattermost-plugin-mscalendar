import {createSelector} from 'reselect';

import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';

import {PluginId} from '../plugin_id';

import {ProviderConfig, ReducerState} from '../reducers';
import {RemoteEvent} from '../types/calendar';

const getPluginState = (state): ReducerState => state['plugins-' + PluginId] || {};

export const getSiteURL = (state): string => {
    const config = getConfig(state);

    let basePath = '';
    if (config && config.SiteURL) {
        basePath = new URL(config.SiteURL).pathname;

        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substring(0, basePath.length - 1);
        }
    }

    return basePath;
};

export const getPluginServerRoute = (state): string => {
    return getSiteURL(state) + '/plugins/' + PluginId;
};

export const getCurrentUserLocale = createSelector(
    getCurrentUser,
    (user) => {
        let locale = 'en';
        if (user && user.locale) {
            locale = user.locale;
        }

        return locale;
    },
);

export const isCreateEventModalVisible = (state) => getPluginState(state).createEventModalVisible;

export const getCreateEventModal = (state) => getPluginState(state).createEventModal;

export const isUserConnected = (state): boolean | null => getPluginState(state).userConnected;

export const getProviderConfiguration = (state): ProviderConfig => getPluginState(state).providerConfiguration;

export function getCalendarEvents(state): RemoteEvent[] {
    const eventsState = getPluginState(state).events;
    if (!eventsState?.activeKey) {
        return [];
    }
    return eventsState.cache?.[eventsState.activeKey] || [];
}

export function getCalendarEventsLoading(state): boolean {
    return getPluginState(state).events?.loading || false;
}

export function getCalendarEventsError(state): string | null {
    return getPluginState(state).events?.error || null;
}
