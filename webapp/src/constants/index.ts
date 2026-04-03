import manifest from '../manifest';

const ActionTypes = {
    CLOSE_CREATE_EVENT_MODAL: `${manifest.id}_close_create_modal`,
    OPEN_CREATE_EVENT_MODAL: `${manifest.id}_open_create_modal`,
    OPEN_CREATE_EVENT_MODAL_WITHOUT_POST: `${manifest.id}_open_create_modal_without_post`,

    RECEIVED_CONNECTED: `${manifest.id}_connected`,
    RECEIVED_DISCONNECTED: `${manifest.id}_disconnected`,
    RECEIVED_PLUGIN_SETTINGS: `${manifest.id}_plugin_settings`,
    RECEIVED_PROVIDER_CONFIGURATION: `${manifest.id}_provider_settings`,

    FETCH_EVENTS_REQUEST: `${manifest.id}_fetch_events_request`,
    RECEIVED_EVENTS: `${manifest.id}_received_events`,
    RECEIVED_CACHED_EVENTS: `${manifest.id}_received_cached_events`,
    FETCH_EVENTS_ERROR: `${manifest.id}_fetch_events_error`,
    RECEIVED_FRESH_EVENTS: `${manifest.id}_received_fresh_events`,
} as const;

export default ActionTypes;
