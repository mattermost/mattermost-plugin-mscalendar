import React, {useCallback, useRef, useState} from 'react';
import {useSelector} from 'react-redux';

import AsyncSelect from 'react-select/async';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';
import {getCurrentTeamId} from 'mattermost-redux/selectors/entities/teams';

import {useAppDispatch} from '@/hooks';
import {getStyleForReactSelect} from '@/utils/styles';
import {autocompleteUserChannels} from '@/actions';

type SelectOption = {
    label: string;
    value: string;
}

type Props = {
    onChange: (selected: string) => void;
    value: string | null;
};

export default function ChannelSelector(props: Props) {
    const [storedError, setStoredError] = useState('');
    const optionCache = useRef<Record<string, string>>({});

    const theme = useSelector(getTheme);
    const teamId = useSelector(getCurrentTeamId);

    const dispatch = useAppDispatch();

    const loadOptions = useCallback(async (input: string): Promise<SelectOption[]> => {
        const response = await dispatch(autocompleteUserChannels(input, teamId));

        if (response.error) {
            setStoredError(response.error);
            return [];
        }

        setStoredError('');

        const options = (response.data ?? []).map((c) => ({
            label: c.display_name,
            value: c.id,
        }));

        for (const opt of options) {
            optionCache.current[opt.value] = opt.label;
        }

        return options;
    }, [dispatch, teamId]);

    const handleChange = (selected: SelectOption | null) => {
        if (selected) {
            optionCache.current[selected.value] = selected.label;
            props.onChange(selected.value);
        }
    };

    return (
        <>
            <AsyncSelect<SelectOption, false>
                value={props.value ? {label: optionCache.current[props.value] || props.value, value: props.value} : null}
                loadOptions={loadOptions}
                defaultOptions={true}
                menuPortalTarget={document.body}
                menuPlacement='auto'
                onChange={handleChange}
                styles={getStyleForReactSelect(theme)}
                isMulti={false}
            />
            {storedError && (
                <div>
                    <span className='error-text'>{storedError}</span>
                </div>
            )}
        </>
    );
}
