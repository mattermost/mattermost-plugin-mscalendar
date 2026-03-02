import manifest from '../manifest';

export const ACTION_TYPES = {
    RECEIVED_EVENTS: `${manifest.id}_received_events`,
    FETCH_EVENTS_REQUEST: `${manifest.id}_fetch_events_request`,
    FETCH_EVENTS_ERROR: `${manifest.id}_fetch_events_error`,
} as const;
