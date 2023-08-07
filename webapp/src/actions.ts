import ActionTypes from './action_types';

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

export const autocompleteConnectedUsers = async (input: string) => {
    return await fetch('/plugins/com.mattermost.gcal/autocomplete/users?search=' + input)
        .then((response) => {
            if (!response.ok) {
                throw new Error("error fetching autocomplete users")
            }
            return response.json();
        })
        .then((data) => {
            return data
        })
        .catch((error) => {
            console.error(error)
        });
}
