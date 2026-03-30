import {combineReducers} from 'redux';

import ActionTypes from '../constants';
import {RemoteEvent} from '../types/calendar';

function userConnected(state: boolean | null = null, action) {
    switch (action.type) {
    case ActionTypes.RECEIVED_CONNECTED:
        return true;
    case ActionTypes.RECEIVED_DISCONNECTED:
        return false;
    default:
        return state;
    }
}

const createEventModalVisible = (state = false, action) => {
    switch (action.type) {
    case ActionTypes.OPEN_CREATE_EVENT_MODAL:
    case ActionTypes.OPEN_CREATE_EVENT_MODAL_WITHOUT_POST:
        return true;
    case ActionTypes.CLOSE_CREATE_EVENT_MODAL:
        return false;
    default:
        return state;
    }
};

const createEventModal = (state = {}, action) => {
    switch (action.type) {
    case ActionTypes.OPEN_CREATE_EVENT_MODAL:
    case ActionTypes.OPEN_CREATE_EVENT_MODAL_WITHOUT_POST:
        return {
            postId: action.data.postId,
            description: action.data.description,
            channelId: action.data.channelId,
            date: action.data.date,
            startTime: action.data.startTime,
            endTime: action.data.endTime,
        };
    case ActionTypes.CLOSE_CREATE_EVENT_MODAL:
        return {};
    default:
        return state;
    }
};

function providerConfiguration(state = null, action) {
    switch (action.type) {
    case ActionTypes.RECEIVED_PROVIDER_CONFIGURATION:
        return action.data;
    default:
        return state;
    }
}

interface EventsState {
    cache: Record<string, RemoteEvent[]>;
    activeKey: string | null;
    activeFrom: string | null;
    activeTo: string | null;
    loading: boolean;
    error: string | null;
}

const eventsInitialState: EventsState = {
    cache: {},
    activeKey: null,
    activeFrom: null,
    activeTo: null,
    loading: false,
    error: null,
};

function events(state: EventsState = eventsInitialState, action: {type: string; data?: RemoteEvent[]; key?: string; from?: string; to?: string; error?: any}): EventsState {
    switch (action.type) {
    case ActionTypes.FETCH_EVENTS_REQUEST:
        return {
            ...state,
            activeKey: action.key || state.activeKey,
            activeFrom: action.from || state.activeFrom,
            activeTo: action.to || state.activeTo,
            loading: true,
            error: null,
        };
    case ActionTypes.RECEIVED_EVENTS: {
        const key = action.key || state.activeKey || '';
        const isLatest = key === state.activeKey;
        return {
            ...state,
            cache: {...state.cache, [key]: action.data || []},
            loading: isLatest ? false : state.loading,
            error: null,
        };
    }
    case ActionTypes.RECEIVED_CACHED_EVENTS:
        return {
            ...state,
            activeKey: action.key || null,
            activeFrom: action.from || state.activeFrom,
            activeTo: action.to || state.activeTo,
            loading: false,
            error: null,
        };
    case ActionTypes.RECEIVED_FRESH_EVENTS: {
        const freshKey = action.key || state.activeKey || '';
        return {
            ...state,
            cache: {[freshKey]: action.data || []},
            activeKey: freshKey,
            activeFrom: action.from || state.activeFrom,
            activeTo: action.to || state.activeTo,
            loading: false,
            error: null,
        };
    }
    case ActionTypes.FETCH_EVENTS_ERROR:
        return {...state, loading: false, error: action.error?.message || 'Failed to fetch events'};
    default:
        return state;
    }
}

export default combineReducers({
    userConnected,
    providerConfiguration,
    createEventModalVisible,
    createEventModal,
    events,
});

export type ProviderFeatures = {
    EncryptedStore: boolean;
    EventNotifications: boolean;
    EnableExperimentalUI: boolean;
}

export type ProviderConfig = {
    Name: string;
    DisplayName: string;
    Repository: string;
    CommandTrigger: string;
    TelemetryShortName: string;
    BotUsername: string;
    BotDisplayName: string;
    Features: ProviderFeatures;
}

export type ReducerState = {
    userConnected: boolean | null;
    createEventModalVisible: boolean;
    createEventModal: {
        channelId?: string;
        postId?: string;
        description?: string;
        date?: string;
        startTime?: string;
        endTime?: string;
    } | null;
    providerConfiguration: ProviderConfig;
    events: {
        cache: Record<string, RemoteEvent[]>;
        activeKey: string | null;
        activeFrom: string | null;
        activeTo: string | null;
        loading: boolean;
        error: string | null;
    };
}
