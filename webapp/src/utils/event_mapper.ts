import {EventInput} from '@fullcalendar/core';

import {RemoteEvent} from '@/types/calendar';

import {getEventStyle, MattermostTheme} from './calendar_theme';

export function mapToFullCalendarEvents(events: RemoteEvent[], theme: MattermostTheme): EventInput[] {
    return events.map((event) => mapToFullCalendarEvent(event, theme));
}

function mapToFullCalendarEvent(event: RemoteEvent, theme: MattermostTheme): EventInput {
    const style = getEventStyle(event, theme);

    return {
        id: event.id,
        title: event.subject || '(No title)',
        start: event.start?.dateTime,
        end: event.end?.dateTime,
        allDay: event.isAllDay || false,
        backgroundColor: style.backgroundColor,
        borderColor: style.borderColor,
        textColor: style.textColor,
        classNames: style.classNames,
        extendedProps: {
            remoteEvent: event,
            borderStyle: style.borderStyle,
            borderWidth: style.borderWidth,
        },
    };
}
