import {Client4} from '@mattermost/client';
import {PostTypes} from 'mattermost-redux/action_types';
import {GlobalState} from '@mattermost/types/store';
import {haveIChannelPermission} from 'mattermost-redux/selectors/entities/roles';
import Permissions from 'mattermost-redux/constants/permissions';
import {Channel} from '@mattermost/types/channels';
import {Dispatch} from 'redux';

import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/common';

import ActionTypes from '../constants';
import {doFetch, doFetchWithResponse} from '../client';
import {PluginId} from '../plugin_id';
import {CreateEventPayload} from '../types/calendar_api_types';
import {RemoteEvent} from '../types/calendar';
import {getPluginServerRoute, getSiteURL} from '../selectors';
import type {ProviderConfig} from '../reducers';

const client = new Client4();

export interface CreateEventPreFill {
    channelId?: string;
    date?: string;
    startTime?: string;
    endTime?: string;
}

export const openCreateEventModal = (channelIdOrPreFill: string | CreateEventPreFill) => {
    const data = typeof channelIdOrPreFill === 'string'
        ? {channelId: channelIdOrPreFill}
        : channelIdOrPreFill;

    return {
        type: ActionTypes.OPEN_CREATE_EVENT_MODAL,
        data,
    };
};

export const closeCreateEventModal = () => {
    return {
        type: ActionTypes.CLOSE_CREATE_EVENT_MODAL,
    };
};

type AutocompleteUser = {
    mm_id: string
    mm_username: string
    mm_display_name: string
}

export type AutocompleteConnectedUsersResponse = {data?: AutocompleteUser[]; error?: string};

export const autocompleteConnectedUsers = (input: string) => async (dispatch, getState): Promise<AutocompleteConnectedUsersResponse> => {
    const state = getState();
    const pluginServerRoute = getPluginServerRoute(state);

    return doFetchWithResponse(`${pluginServerRoute}/autocomplete/users?search=${input}`).
        then((response) => {
            return {data: response.data};
        }).
        catch((response) => {
            const error = response.message?.error || 'An error occurred while searching for users.';
            return {data: [], error};
        });
};

export type AutocompleteChannelsResponse = {data?: Channel[]; error?: string};

export const autocompleteUserChannels = (input: string, teamId: string) => async (dispatch, getState): Promise<AutocompleteChannelsResponse> => {
    const state = getState();
    const siteURL = getSiteURL(state);
    client.setUrl(siteURL);

    try {
        const channels = await client.autocompleteChannels(teamId, input);
        const channelsCanWriteTo = channels.filter((c) => haveIChannelPermission(state, teamId, c.id, Permissions.CREATE_POST));
        return {data: channelsCanWriteTo};
    } catch (e: any) {
        const error = e.message?.error || 'An error occurred while searching for channels.';
        return {data: [], error};
    }
};

export type CreateCalendarEventResponse = {data?: any; error?: string};

export const createCalendarEvent = (payload: CreateEventPayload) => async (dispatch, getState): Promise<CreateCalendarEventResponse> => {
    const state = getState();
    const pluginServerRoute = getPluginServerRoute(state);

    return doFetchWithResponse(`${pluginServerRoute}/api/v1/events/create`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
    }).
        then((data) => {
            return {data};
        }).
        catch((response) => {
            const error = response.message?.error || 'An error occurred while creating the event.';
            return {error};
        });
};

export function getConnected() {
    return async (dispatch, getState) => {
        let data;
        const baseUrl = getPluginServerRoute(getState());
        try {
            data = await doFetch(`${baseUrl}/api/v1/me`, {
                method: 'get',
            });
        } catch (error) {
            dispatch({type: ActionTypes.RECEIVED_DISCONNECTED});
            return {error};
        }

        dispatch({
            type: ActionTypes.RECEIVED_CONNECTED,
            data,
        });

        return {data};
    };
}

export function sendEphemeralPost(message: string, channelId?: string) {
    return (dispatch, getState) => {
        const timestamp = Date.now();
        const post = {
            id: 'mscalplugin_' + Date.now(),
            user_id: getState().entities.users.currentUserId,
            channel_id: channelId || getCurrentChannelId(getState()),
            message,
            type: 'system_ephemeral',
            create_at: timestamp,
            update_at: timestamp,
            root_id: '',
            parent_id: '',
            props: {},
        };

        dispatch({
            type: PostTypes.RECEIVED_NEW_POST,
            data: post,
            channelId,
        });
    };
}

export function handleConnect(store) {
    return (msg) => {
        store.dispatch({
            type: ActionTypes.RECEIVED_CONNECTED,
            data: msg.data,
        });
    };
}

export function handleDisconnect(store) {
    return (msg) => {
        store.dispatch({
            type: ActionTypes.RECEIVED_DISCONNECTED,
            data: msg.data,
        });
    };
}

export function getProviderConfiguration() {
    return async (dispatch, getState): Promise<ProviderConfig | null> => {
        let data;
        const baseUrl = getPluginServerRoute(getState());
        try {
            data = await doFetch(`${baseUrl}/api/v1/provider`, {
                method: 'get',
            });

            dispatch({
                type: ActionTypes.RECEIVED_PROVIDER_CONFIGURATION,
                data,
            });
        } catch (error) {
            return {error};
        }

        return data;
    };
}

export const refreshCalendarEvents = (from: string, to: string) => async (dispatch: Dispatch, getState: () => GlobalState) => {
    const key = makeEventsCacheKey(from, to);
    dispatch({type: ActionTypes.FETCH_EVENTS_REQUEST, key});

    const pluginServerRoute = getPluginServerRoute(getState());
    const params = new URLSearchParams({from, to});

    try {
        const freshEvents: RemoteEvent[] = await doFetch(`${pluginServerRoute}/api/v1/events/view?${params.toString()}`, {
            method: 'get',
        });
        dispatch({type: ActionTypes.RECEIVED_FRESH_EVENTS, data: freshEvents, key});
        return {data: freshEvents, error: null};
    } catch (error) {
        dispatch({type: ActionTypes.FETCH_EVENTS_ERROR, error});
        return {data: null, error};
    }
};

export const refreshActiveCalendarView = () => async (dispatch: Dispatch, getState: () => GlobalState) => {
    const state = getState() as any;
    const pluginState = state['plugins-' + PluginId];
    const activeKey = pluginState?.events?.activeKey;
    if (!activeKey) {
        return;
    }

    const [fromDate, toDate] = activeKey.split('|');
    const from = `${fromDate}T00:00:00.000Z`;
    const to = `${toDate}T00:00:00.000Z`;
    await (dispatch as any)(refreshCalendarEvents(from, to));
};

function makeEventsCacheKey(from: string, to: string): string {
    return `${from.split('T')[0]}|${to.split('T')[0]}`;
}

export const fetchCalendarEvents = (from: string, to: string) => async (dispatch: Dispatch, getState: () => GlobalState) => {
    const key = makeEventsCacheKey(from, to);
    const state = getState() as any;
    const pluginState = state['plugins-' + PluginId];
    const cached = pluginState?.events?.cache?.[key];

    if (cached) {
        dispatch({type: ActionTypes.RECEIVED_CACHED_EVENTS, key});
        return {data: cached, error: null};
    }

    dispatch({type: ActionTypes.FETCH_EVENTS_REQUEST, key});

    const pluginServerRoute = getPluginServerRoute(state);
    const params = new URLSearchParams({from, to});

    try {
        const events: RemoteEvent[] = await doFetch(`${pluginServerRoute}/api/v1/events/view?${params.toString()}`, {
            method: 'get',
        });
        dispatch({type: ActionTypes.RECEIVED_EVENTS, data: events, key});
        return {data: events, error: null};
    } catch (error) {
        dispatch({type: ActionTypes.FETCH_EVENTS_ERROR, error});
        return {data: null, error};
    }
};
