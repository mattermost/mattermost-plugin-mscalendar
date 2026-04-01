import {useEffect, useRef} from 'react';
import {createPortal} from 'react-dom';

import {RemoteEvent} from '@/types/calendar';
import {MattermostTheme} from '@/utils/calendar_theme';

import './event_tooltip.scss';

interface EventTooltipProps {
    event: RemoteEvent;
    anchorRect: DOMRect;
    timezone: string;
    theme: MattermostTheme;
    onClose: () => void;
}

function formatResponseStatus(response?: string): {label: string; className: string} {
    switch (response) {
    case 'accepted':
        return {label: 'Accepted', className: 'mscalendar-tooltip__status--accepted'};
    case 'tentative':
    case 'tentativelyAccepted':
        return {label: 'Tentative', className: 'mscalendar-tooltip__status--tentative'};
    case 'declined':
        return {label: 'Declined', className: 'mscalendar-tooltip__status--declined'};
    case 'not_answered':
    case 'notResponded':
    case 'needsAction':
    default:
        return {label: 'Not responded', className: 'mscalendar-tooltip__status--pending'};
    }
}

function formatEventTime(event: RemoteEvent, timezone: string): string {
    if (event.isAllDay) {
        return 'All day';
    }

    const tz = timezone || 'UTC';
    const options: Intl.DateTimeFormatOptions = {
        hour: 'numeric',
        minute: '2-digit',
        timeZone: tz,
    };
    const dateOptions: Intl.DateTimeFormatOptions = {
        ...options,
        day: 'numeric',
        month: 'short',
        year: 'numeric',
    };

    const start = event.start?.dateTime ? new Date(event.start.dateTime) : null;
    const end = event.end?.dateTime ? new Date(event.end.dateTime) : null;

    if (!start) {
        return '';
    }

    const startDate = start.toLocaleDateString([], {day: 'numeric', month: 'short', year: 'numeric', timeZone: tz});
    const startTime = start.toLocaleTimeString([], options);

    if (!end) {
        return `${startDate} ${startTime}`;
    }

    const endDate = end.toLocaleDateString([], {day: 'numeric', month: 'short', year: 'numeric', timeZone: tz});
    const endTime = end.toLocaleTimeString([], options);

    if (startDate === endDate) {
        return `${startDate} ${startTime} - ${endTime}`;
    }

    return `${start.toLocaleDateString([], dateOptions)} - ${end.toLocaleDateString([], dateOptions)}`;
}

const EventTooltip = ({event, anchorRect, timezone, theme, onClose}: EventTooltipProps) => {
    const tooltipRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleEsc = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                onClose();
            }
        };
        document.addEventListener('keydown', handleEsc);
        return () => document.removeEventListener('keydown', handleEsc);
    }, [onClose]);

    const spaceBelow = window.innerHeight - anchorRect.bottom;
    const fitsBelow = spaceBelow > 200;

    const verticalPosition = fitsBelow ?
        {top: anchorRect.bottom + 4} :
        {bottom: (window.innerHeight - anchorRect.top) + 4};
    const right = Math.max(8, window.innerWidth - anchorRect.right);

    const status = formatResponseStatus(event.responseStatus?.response);
    const timeDisplay = formatEventTime(event, timezone);
    const locationName = event.location?.displayName;
    const conferenceUrl = event.conference?.url;
    const conferenceName = event.conference?.application;
    const organizerName = event.organizer?.emailAddress?.name || event.organizer?.emailAddress?.address;
    const weblink = event.weblink;

    const tooltip = (
        <div
            className='mscalendar-tooltip__backdrop'
            onClick={onClose}
        >
            <div
                ref={tooltipRef}
                className='mscalendar-tooltip'
                onClick={(e) => e.stopPropagation()}
                style={{
                    ...verticalPosition,
                    right,
                    backgroundColor: theme.centerChannelBg,
                    color: theme.centerChannelColor,
                }}
            >
                <div className='mscalendar-tooltip__header'>
                    <a
                        className='mscalendar-tooltip__title'
                        href={weblink}
                        target='_blank'
                        rel='noopener noreferrer'
                        style={{color: theme.linkColor}}
                    >
                        {event.subject || '(No title)'}
                        <i
                            className='icon icon-open-in-new'
                            style={{fontSize: '14px', marginLeft: '4px'}}
                        />
                    </a>
                    <button
                        type='button'
                        className='mscalendar-tooltip__close'
                        onClick={onClose}
                        style={{color: theme.centerChannelColor}}
                    >
                        <i className='icon icon-close'/>
                    </button>
                </div>

                <div
                    className='mscalendar-tooltip__time'
                    style={{color: theme.centerChannelColor}}
                >
                    <i className='icon icon-clock-outline'/>
                    <span>{timeDisplay}</span>
                </div>

                {locationName && (
                    <div className='mscalendar-tooltip__row'>
                        <i className='icon icon-map-marker-outline'/>
                        <span>{locationName}</span>
                    </div>
                )}

                {conferenceUrl && (
                    <div className='mscalendar-tooltip__conference'>
                        <a
                            className='mscalendar-tooltip__join-btn'
                            href={conferenceUrl}
                            target='_blank'
                            rel='noopener noreferrer'
                            style={{
                                backgroundColor: theme.buttonBg,
                                color: theme.buttonColor,
                            }}
                        >
                            {'Join'}
                        </a>
                        {conferenceName && (
                            <span
                                className='mscalendar-tooltip__conference-name'
                                style={{color: theme.centerChannelColor}}
                            >
                                <i className='icon icon-video-outline'/>
                                {conferenceName}
                            </span>
                        )}
                    </div>
                )}

                {organizerName && (
                    <div className='mscalendar-tooltip__row'>
                        <i className='icon icon-account-outline'/>
                        <div className='mscalendar-tooltip__organizer'>
                            <span>{organizerName}</span>
                            <span className='mscalendar-tooltip__organizer-label'>{'Organizer'}</span>
                        </div>
                    </div>
                )}

                <div className='mscalendar-tooltip__row mscalendar-tooltip__row--status'>
                    <i className='icon icon-check-circle-outline'/>
                    <span className={`mscalendar-tooltip__status ${status.className}`}>
                        {status.label}
                    </span>
                </div>
            </div>
        </div>
    );

    return createPortal(tooltip, document.body);
};

export default EventTooltip;
