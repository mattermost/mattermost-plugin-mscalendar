import React, {useMemo} from 'react';
import {useSelector} from 'react-redux';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import {getTodayString} from '@/utils/datetime';
import {CreateEventPayload} from '@/types/calendar_api_types';

import ReactSelectSetting from './react_select_setting';

const minuteStep = 15;

type Props = {
    value: string;
    onChange: (name: keyof CreateEventPayload, value: string) => void;
    startTime?: string
    endTime?: string
    date?: string
}

type Option = {
    label: string
    value: string
}

export default function TimeSelector(props: Props) {
    const theme = useSelector(getTheme);

    const isEndTimeSelector = Boolean(props.startTime);
    const isStartTimeSelector = !isEndTimeSelector;

    const options: Option[] = useMemo(() => {
        let fromHour = 0;
        let fromMinute = 0;
        let toHour = 23;
        let toMinute = isStartTimeSelector ? 30 : 45;
        let ranges: string[] = [];

        if (props.date === getTodayString()) {
            const now = new Date();
            const roundedMinutes = Math.ceil(now.getMinutes() / minuteStep) * minuteStep;
            fromHour = now.getHours() + Math.floor(roundedMinutes / 60);
            fromMinute = roundedMinutes % 60;
            ranges = generateMilitaryTimeArray(fromHour, fromMinute, toHour, toMinute);
        }

        if (props.startTime) {
            const parsed = parseHHMM(props.startTime);
            fromHour = parsed.hour;
            fromMinute = parsed.minute + minuteStep;
            const extraHours = Math.floor(fromMinute / 60);
            fromMinute %= 60;
            fromHour += extraHours;
            if (fromHour < 24) {
                ranges = generateMilitaryTimeArray(fromHour, fromMinute, toHour, toMinute);
            }
        }

        if (props.endTime) {
            const parsed = parseHHMM(props.endTime);
            toHour = parsed.hour;
            toMinute = parsed.minute;
            if (isStartTimeSelector) {
                const endTotal = (toHour * 60) + toMinute;
                const maxStartTotal = endTotal - minuteStep;
                if (maxStartTotal < 0) {
                    return [];
                }
                toHour = Math.floor(maxStartTotal / 60);
                toMinute = maxStartTotal % 60;
            }
            ranges = generateMilitaryTimeArray(fromHour, fromMinute, toHour, toMinute);
        }

        if (!ranges.length && !props.startTime) {
            ranges = generateMilitaryTimeArray(0, 0, toHour, toMinute);
        }

        return ranges.map((t) => ({
            label: t,
            value: t,
        }));
    }, [props.startTime, props.endTime, props.date, isStartTimeSelector]);

    let value: Option | undefined | null;
    if (props.value) {
        value = options.find((option: Option) => option.value === props.value);
    }

    const handleChange = (_: string | undefined, newValue: string | string[] | null) => {
        const selectedTime = typeof newValue === 'string' ? newValue : null;
        if (!selectedTime) {
            return;
        }
        if (props.startTime) {
            props.onChange('end_time', selectedTime);
        } else {
            props.onChange('start_time', selectedTime);

            options.forEach((option: Option, i: number) => {
                if (option.value === selectedTime) {
                    props.onChange('end_time', i + 2 < options.length ? options[i + 2].value : '');
                }
            });
        }
    };

    return (
        <ReactSelectSetting
            value={value}
            onChange={handleChange}
            theme={theme}
            options={options}
        />
    );
}

const parseHHMM = (time: string): {hour: number; minute: number} => {
    const parts = time.split(':');
    const hour = parseInt(parts[0], 10);
    const minute = parseInt(parts[1], 10);
    return {
        hour: Number.isNaN(hour) ? 0 : hour,
        minute: Number.isNaN(minute) ? 0 : minute,
    };
};

const generateMilitaryTimeArray = (fromHour = 0, fromMinute = 0, toHour = 23, toMinute = 45, step = minuteStep) => {
    const timeArray = [];
    for (let hour = fromHour; hour <= toHour; hour++) {
        const startMinute = hour === fromHour ? fromMinute : 0;
        const endMinute = hour === toHour ? toMinute : 45;
        for (let minute = startMinute; minute <= endMinute; minute += step) {
            const formattedHour = hour.toString().padStart(2, '0');
            const formattedMinute = minute.toString().padStart(2, '0');
            const timeString = `${formattedHour}:${formattedMinute}`;
            timeArray.push(timeString);
        }
    }
    return timeArray;
};
