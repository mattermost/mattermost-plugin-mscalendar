import {Dispatch} from 'redux';

import client from '@/client/client';
import {ACTION_TYPES} from '@/constants';

export const fetchCalendarEvents = (from: string, to: string) => async (dispatch: Dispatch) => {
    dispatch({type: ACTION_TYPES.FETCH_EVENTS_REQUEST});

    try {
        const events = await client.getCalendarEvents(from, to);
        dispatch({type: ACTION_TYPES.RECEIVED_EVENTS, data: events});
        return {data: events, error: null};
    } catch (error) {
        dispatch({type: ACTION_TYPES.FETCH_EVENTS_ERROR, error});
        return {data: null, error};
    }
};
