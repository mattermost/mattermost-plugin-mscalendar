import React, {useCallback} from 'react';
import {useSelector} from 'react-redux';

import AsyncSelect from 'react-select/async';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import {getStyleForReactSelect} from '@/utils/styles';
import {autocompleteConnectedUsers} from '@/actions';

type SelectOption = {
    label: string;
    value: string;
}

type Props = {
    onChange: (selected: string[]) => void;
    value: string[];
};

export default function AttendeeSelector(props: Props) {
    const theme = useSelector(getTheme);

    const loadOptions = useCallback(async (input: string): Promise<SelectOption[]> => {
        const matchedUsers = await autocompleteConnectedUsers(input);

        return matchedUsers.map(u => ({
            label: u,
            value: u,
        }));
    }, []);

    const handleChange = (selected: SelectOption[]) => {
        props.onChange(selected.map(option => option.value));
    }

    return (
        <AsyncSelect
            value={props.value}
            loadOptions={loadOptions}
            defaultOptions={true}
            menuPortalTarget={document.body}
            menuPlacement='auto'
            onChange={handleChange}
            styles={getStyleForReactSelect(theme)}
            isMulti={true}
        />
    );
}
