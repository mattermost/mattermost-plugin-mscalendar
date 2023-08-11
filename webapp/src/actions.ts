import ActionTypes from './action_types';
import {doFetchWithResponse} from './client';
import {PluginId} from './plugin_id';
import {CreateEventPayload} from './types/calendar_api_types';

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

type AutocompleteUser = {
    mm_id: string
    mm_username: string
    mm_display_name: string
}

export const autocompleteConnectedUsers = async (input: string): Promise<{data: AutocompleteUser[]} | {error: string}> => {
    return doFetchWithResponse(`/plugins/${PluginId}/autocomplete/users?search=${input}`).
        then((response) => {
            return {data: response.data};
        }).
        catch(response => {
            if ('message' in response && 'error' in response.message) {
                return response.message;
            }

            return {error: 'An error occurred while searching for users.'};
        });
};

type AutocompleteChannel = {
    id: string
    display_name: string
}

export const autocompleteUserChannels = async (input: string): Promise<{data: AutocompleteChannel[]} | {error: string}> => {
    return doFetchWithResponse(`/plugins/${PluginId}/autocomplete/channels?search=${input}`).
        then((response) => {
            return {data: response.data};
        }).
        catch(response => {
            if ('message' in response && 'error' in response.message) {
                return response.message;
            }

            return {error: 'An error occurred while searching for channels.'};
        });
};

export const createCalendarEvent = async (payload: CreateEventPayload): Promise<{error: string} | {data: any}> => {
    return doFetchWithResponse('/plugins/com.mattermost.gcal/api/v1/events/create', {
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
            if ('message' in response && 'error' in response.message) {
                return response.message;
            }

            return {error: 'An error occurred while creating the event.'};
        });
};
