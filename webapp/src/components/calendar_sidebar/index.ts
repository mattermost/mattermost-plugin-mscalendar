import {connect} from 'react-redux';
import {bindActionCreators, Dispatch} from 'redux';
import {GlobalState} from '@mattermost/types/store';
import {getCurrentTimezone} from 'mattermost-redux/selectors/entities/timezone';
import {getBool} from 'mattermost-redux/selectors/entities/preferences';
import Preferences from 'mattermost-redux/constants/preferences';

import {fetchCalendarEvents, refreshCalendarEvents, getConnected, sendEphemeralPost, openCreateEventModal} from '@/actions';
import {getCalendarEvents, getCalendarEventsLoading, getCalendarEventsError, isUserConnected, getPluginServerRoute} from '@/selectors';

import CalendarSidebar from './calendar_sidebar';

export function mapStateToProps(state: GlobalState) {
    return {
        events: getCalendarEvents(state),
        loading: getCalendarEventsLoading(state),
        error: getCalendarEventsError(state),
        timezone: getCurrentTimezone(state),
        connected: isUserConnected(state),
        pluginServerRoute: getPluginServerRoute(state),
        useMilitaryTime: getBool(state, Preferences.CATEGORY_DISPLAY_SETTINGS, Preferences.USE_MILITARY_TIME, false),
    };
}

function mapDispatchToProps(dispatch: Dispatch) {
    return {
        actions: bindActionCreators({
            fetchCalendarEvents,
            refreshCalendarEvents,
            getConnected,
            sendEphemeralPost,
            openCreateEventModal,
        }, dispatch),
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(CalendarSidebar);
