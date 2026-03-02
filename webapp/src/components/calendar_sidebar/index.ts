import {connect} from 'react-redux';
import {bindActionCreators, Dispatch} from 'redux';
import {GlobalState} from '@mattermost/types/store';
import {getCurrentTimezone} from 'mattermost-redux/selectors/entities/timezone';

import {fetchCalendarEvents} from '@/actions';
import {getCalendarEvents, getCalendarEventsLoading, getCalendarEventsError} from '@/selectors';

import CalendarSidebar from './calendar_sidebar';

function mapStateToProps(state: GlobalState) {
    return {
        events: getCalendarEvents(state),
        loading: getCalendarEventsLoading(state),
        error: getCalendarEventsError(state),
        timezone: getCurrentTimezone(state),
    };
}

function mapDispatchToProps(dispatch: Dispatch) {
    return {
        actions: bindActionCreators({
            fetchCalendarEvents,
        }, dispatch),
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(CalendarSidebar);
