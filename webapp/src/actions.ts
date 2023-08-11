import ActionTypes from './action_types';
import {doFetchWithResponse} from './client';
import { PluginId } from './plugin_id';

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

export const autocompleteConnectedUsers = async (input: string): Promise<AutocompleteUser[]> => {
    return doFetchWithResponse(`/plugins/${PluginId}/autocomplete/users?search=` + input).
        then((response) => {
            if (!response.response.ok) {
                throw new Error('error fetching autocomplete users');
            }
            return response.data;
        }).
        then((data) => {
            return data;
        }).
        catch((error) => {
            throw new Error(error);
        });
};

type AutocompleteChannel = {
    id: string
    display_name: string
}

export const autocompleteUserChannels = async (input: string): Promise<AutocompleteChannel[]> => {
    return doFetchWithResponse(`/plugins/${PluginId}/autocomplete/users?search=` + input).
        then((response) => {
            if (!response.response.ok) {
                throw new Error('error fetching autocomplete channels');
            }
            return response.data;
        }).
        then((data) => {
            return data;
        }).
        catch((error) => {
            throw new Error(error);
        });
};
