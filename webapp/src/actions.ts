import Client4 from 'mattermost-redux/client/client4';
import {GlobalState} from 'mattermost-redux/types/store';
import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {haveIChannelPermission} from 'mattermost-redux/selectors/entities/roles';
import Permissions from 'mattermost-redux/constants/permissions';
import {Channel} from '@mattermost/types/lib/channels';

import ActionTypes from './action_types';
import {doFetchWithResponse} from './client';
import {PluginId} from './plugin_id';
import {CreateEventPayload} from './types/calendar_api_types';

const client = new Client4();

export const openCreateEventModal = (channelId: string) => {
    return {
        type: ActionTypes.OPEN_CREATE_EVENT_MODAL,
        data: {
            channelId,
        },
    };
};

export const closeCreateEventModal = () => {
    return {
        type: ActionTypes.CLOSE_CREATE_EVENT_MODAL,
    };
};

export const getSiteURL = (state: GlobalState): string => {
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

export const getPluginServerRoute = (state: GlobalState): string => {
    const siteURL = getSiteURL(state);
    return `${siteURL}/plugins/${PluginId}`;
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
        const channelsCanWriteTo = channels.filter((c) => haveIChannelPermission(state, {channel: c.id, permission: Permissions.CREATE_POST}));
        return {data: channelsCanWriteTo};
    } catch (e) {
        const error = response.message?.error || 'An error occurred while searching for channels.';
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
