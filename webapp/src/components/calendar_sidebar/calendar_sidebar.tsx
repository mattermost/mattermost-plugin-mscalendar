import {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import FullCalendar from '@fullcalendar/react';
import timeGridPlugin from '@fullcalendar/timegrid';

import {RemoteEvent} from '@/types/calendar';
import {getContainerStyle, MattermostTheme} from '@/utils/calendar_theme';
import {mapToFullCalendarEvents} from '@/utils/event_mapper';

import './calendar_sidebar.scss';

interface CalendarSidebarProps {
    theme: MattermostTheme;
    events: RemoteEvent[];
    loading: boolean;
    error: string | null;
    timezone: string;
    actions: {
        fetchCalendarEvents: (from: string, to: string) => void;
    };
}

const EXPANDED_WIDTH_THRESHOLD = 500;

function getDateRange(isWeekView: boolean): {from: string; to: string} {
    const now = new Date();
    const start = new Date(now.getFullYear(), now.getMonth(), now.getDate());

    let end: Date;
    if (isWeekView) {
        end = new Date(start);
        end.setDate(end.getDate() + 7);
    } else {
        end = new Date(start);
        end.setDate(end.getDate() + 1);
    }

    return {
        from: start.toISOString(),
        to: end.toISOString(),
    };
}

function getCurrentScrollTime(): string {
    const now = new Date();
    const hours = Math.max(0, now.getHours() - 1);
    return `${String(hours).padStart(2, '0')}:00:00`;
}

const CalendarSidebar = ({theme, events, loading, error, timezone, actions}: CalendarSidebarProps) => {
    const calendarRef = useRef<FullCalendar>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [isExpanded, setIsExpanded] = useState(false);

    const fetchEvents = useCallback((weekView: boolean) => {
        const {from, to} = getDateRange(weekView);
        actions.fetchCalendarEvents(from, to);
    }, [actions]);

    useEffect(() => {
        fetchEvents(isExpanded);
    }, [fetchEvents, isExpanded]);

    useEffect(() => {
        const container = containerRef.current;
        if (!container) {
            return () => { /* noop */ };
        }

        const observer = new ResizeObserver((entries) => {
            for (const entry of entries) {
                const width = entry.contentRect.width;
                setIsExpanded(width >= EXPANDED_WIDTH_THRESHOLD);
            }
        });

        observer.observe(container);
        return () => observer.disconnect();
    }, []);

    useEffect(() => {
        const api = calendarRef.current?.getApi();
        if (!api) {
            return;
        }

        const targetView = isExpanded ? 'timeGridWeek' : 'timeGridDay';
        if (api.view.type !== targetView) {
            api.changeView(targetView);
        }
    }, [isExpanded]);

    const fullCalendarEvents = useMemo(
        () => mapToFullCalendarEvents(events, theme),
        [events, theme],
    );

    const containerStyle = useMemo(() => getContainerStyle(theme), [theme]);
    const scrollTime = useMemo(() => getCurrentScrollTime(), []);

    return (
        <div
            ref={containerRef}
            className='mscalendar-sidebar'
            style={containerStyle}
        >
            {error && (
                <div style={{padding: '8px 0', color: theme.dndIndicator || '#D24B4E', fontSize: '0.85em'}}>
                    {'Unable to load calendar events. Please ensure your account is connected.'}
                </div>
            )}
            <div className='mscalendar-sidebar__calendar'>
                <FullCalendar
                    ref={calendarRef}
                    plugins={[timeGridPlugin]}
                    initialView='timeGridDay'
                    timeZone={timezone || 'local'}
                    nowIndicator={true}
                    headerToolbar={false}
                    allDaySlot={false}
                    slotDuration='00:30:00'
                    scrollTime={scrollTime}
                    height='100%'
                    events={fullCalendarEvents}
                    eventDisplay='block'
                    slotEventOverlap={false}
                    expandRows={true}
                />
            </div>
        </div>
    );
};

export default CalendarSidebar;
