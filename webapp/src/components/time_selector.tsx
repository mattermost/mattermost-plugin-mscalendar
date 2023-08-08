import React, {useMemo} from 'react';
import {useSelector} from 'react-redux';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import ReactSelectSetting from './react_select_setting';

type Props = {
    value: string;
    onChange: (value: string) => void;

    // TODO: implement upper bound and lower bound to make sure end > start
    // upperBound: idk;
    // lowerBound: idk;
}

export default function TimeSelector(props: Props) {
    const theme = useSelector(getTheme);

    const options = useMemo(() => militaryTimeOptions.map((t) => ({
        label: t,
        value: t,
    })), []);

    let value = null;
    if (props.value) {
        value = options.find((option) => option.value === props.value);
    }

    return (
        <ReactSelectSetting
            value={value}
            onChange={(_, time) => {
                props.onChange(time);
            }}
            theme={theme}
            options={options}
        />
    );
}

const generateMilitaryTimeArray = (fromHour = 0, fromMinute = 0, toHour = 0, toMinute = 0, step = 15) => {
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

const militaryTimeOptions = generateMilitaryTimeArray();
