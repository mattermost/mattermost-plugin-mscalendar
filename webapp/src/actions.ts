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
