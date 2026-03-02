import {EventInput} from '@fullcalendar/core';

import {RemoteEvent} from '@/types/calendar';

import {getShowAsColor, MattermostTheme} from './calendar_theme';

export function mapToFullCalendarEvents(events: RemoteEvent[], theme: MattermostTheme): EventInput[] {
    return events.map((event) => mapToFullCalendarEvent(event, theme));
}

function mapToFullCalendarEvent(event: RemoteEvent, theme: MattermostTheme): EventInput {
    return {
        id: event.id,
        title: event.subject || '(No title)',
        start: event.start?.dateTime,
        end: event.end?.dateTime,
        allDay: event.isAllDay || false,
        backgroundColor: getShowAsColor(event.showAs, theme),
        borderColor: getShowAsColor(event.showAs, theme),
        extendedProps: {
            showAs: event.showAs,
            location: event.location?.displayName,
            conference: event.conference?.url,
            weblink: event.weblink,
            importance: event.importance,
            isOrganizer: event.isOrganizer,
            isCancelled: event.isCancelled,
            responseStatus: event.responseStatus?.response,
        },
    };
}
