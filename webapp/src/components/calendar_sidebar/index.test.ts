import {GlobalState} from '@mattermost/types/store';
import Preferences from 'mattermost-redux/constants/preferences';

jest.mock('./calendar_sidebar', () => () => null);

import {mapStateToProps} from './index';

describe('CalendarSidebar container mapStateToProps', () => {
    const baseState = {
        entities: {
            users: {
                currentUserId: 'user_1',
                profiles: {
                    user_1: {
                        id: 'user_1',
                        timezone: {
                            useAutomaticTimezone: false,
                            manualTimezone: 'UTC',
                            automaticTimezone: '',
                        },
                    },
                },
            },
            general: {
                config: {
                    SiteURL: 'http://localhost:8065',
                },
            },
            preferences: {
                myPreferences: {},
            },
            timezone: {
                userTimezone: {
                    useAutomaticTimezone: false,
                    manualTimezone: 'UTC',
                    automaticTimezone: '',
                },
            },
        },
        'plugins-com.mattermost.mscalendar': {
            connected: true,
            calendarEvents: [],
            calendarEventsLoading: false,
            calendarEventsError: null,
            pluginServerRoute: '/plugins/com.mattermost.mscalendar',
        },
    } as unknown as GlobalState;

    it('returns useMilitaryTime: false when preference is absent', () => {
        const props = mapStateToProps(baseState);
        expect(props.useMilitaryTime).toBe(false);
    });

    it('returns useMilitaryTime: true when preference is true', () => {
        const state = {
            ...baseState,
            entities: {
                ...baseState.entities,
                preferences: {
                    myPreferences: {
                        [`${Preferences.CATEGORY_DISPLAY_SETTINGS}--${Preferences.USE_MILITARY_TIME}`]: {
                            category: Preferences.CATEGORY_DISPLAY_SETTINGS,
                            name: Preferences.USE_MILITARY_TIME,
                            value: 'true',
                        },
                    },
                },
            },
        } as unknown as GlobalState;

        const props = mapStateToProps(state);
        expect(props.useMilitaryTime).toBe(true);
    });

    it('returns useMilitaryTime: false when preference is false', () => {
        const state = {
            ...baseState,
            entities: {
                ...baseState.entities,
                preferences: {
                    myPreferences: {
                        [`${Preferences.CATEGORY_DISPLAY_SETTINGS}--${Preferences.USE_MILITARY_TIME}`]: {
                            category: Preferences.CATEGORY_DISPLAY_SETTINGS,
                            name: Preferences.USE_MILITARY_TIME,
                            value: 'false',
                        },
                    },
                },
            },
        } as unknown as GlobalState;

        const props = mapStateToProps(state);
        expect(props.useMilitaryTime).toBe(false);
    });
});
