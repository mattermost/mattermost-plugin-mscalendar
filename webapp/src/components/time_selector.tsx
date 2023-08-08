import React, {useEffect, useMemo} from 'react';
import {useSelector} from 'react-redux';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import ReactSelectSetting from './react_select_setting';

const minuteStep = 15;

type Props = {
    value: string;
    onChange: (value: string) => void;
    startTime?: string
    endTime?: string
}

export default function TimeSelector(props: Props) {
    const theme = useSelector(getTheme);
    let options = null;
    let value = null;
    let ranges: string[];

    const updateOptions = () => {
        let fromHour, fromMinute = 0;
        let toHour = 23;
        let toMinute = 45;

        if (props.startTime != undefined && props.startTime != '') {
            const parts = props.startTime.split(":")
            fromHour = parseInt(parts[0]);
            fromMinute = parseInt(parts[1]) + minuteStep;
            ranges = generateMilitaryTimeArray(fromHour, fromMinute, toHour, toMinute)
        }

        if (props.endTime != undefined && props.endTime != '') {
            const parts = props.endTime.split(":")
            toHour = parseInt(parts[0]);
            toMinute = parseInt(parts[1]);
            ranges = generateMilitaryTimeArray(fromHour, fromMinute, toHour, toMinute)
            console.log("to", toHour, toMinute)
        }

        if (ranges == undefined) {
            ranges = generateMilitaryTimeArray()
        }

        options = ranges.map((t) => ({
            label: t,
            value: t,
        }))

        if (props.value) {
            value = options.find((option) => option.value === props.value);
        }
    }

    const handleChange = (_, time) => {
        console.log(time)
        props.onChange(time);
    }

    useEffect(updateOptions, [props.startTime, props.endTime])
    updateOptions()

    return (
        <ReactSelectSetting
            value={value}
            onChange={handleChange}
            theme={theme}
            options={options}
        />
    );
}

const generateMilitaryTimeArray = (fromHour = 0, fromMinute = 0, toHour = 23, toMinute = 45, step = minuteStep) => {
    const timeArray = [];
    for (let hour = fromHour; hour <= toHour; hour++) {
        if (hour != fromHour) fromMinute = 0
        if (hour != toHour) toMinute = 45
        for (let minute = fromMinute; minute <= toMinute; minute += step) {
            const formattedHour = hour.toString().padStart(2, '0');
            const formattedMinute = minute.toString().padStart(2, '0');
            const timeString = `${formattedHour}:${formattedMinute}`;
            timeArray.push(timeString);
        }
    }
    return timeArray;
};
