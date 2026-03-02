import {combineReducers} from 'redux';

import {RemoteEvent} from '@/types/calendar';
import {ACTION_TYPES} from '@/constants';

interface EventsState {
    items: RemoteEvent[];
    loading: boolean;
    error: string | null;
}

const initialState: EventsState = {
    items: [],
    loading: false,
    error: null,
};

function events(state: EventsState = initialState, action: {type: string; data?: RemoteEvent[]; error?: any}): EventsState {
    switch (action.type) {
    case ACTION_TYPES.FETCH_EVENTS_REQUEST:
        return {...state, loading: true, error: null};
    case ACTION_TYPES.RECEIVED_EVENTS:
        return {...state, items: action.data || [], loading: false, error: null};
    case ACTION_TYPES.FETCH_EVENTS_ERROR:
        return {...state, loading: false, error: action.error?.message || 'Failed to fetch events'};
    default:
        return state;
    }
}

export default combineReducers({
    events,
});
