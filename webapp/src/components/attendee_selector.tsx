import React, {useCallback, useRef, useState} from 'react';
import {useSelector} from 'react-redux';

import AsyncCreatableSelect from 'react-select/async-creatable';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import {useAppDispatch} from '@/hooks';
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
    const [storedError, setStoredError] = useState('');
    const optionCache = useRef<Record<string, string>>({});

    const theme = useSelector(getTheme);

    const dispatch = useAppDispatch();

    const loadOptions = useCallback(async (input: string): Promise<SelectOption[]> => {
        const response = await dispatch(autocompleteConnectedUsers(input));

        if (response.error) {
            setStoredError(response.error);
            return [];
        }

        setStoredError('');

        const options = (response.data ?? []).map((u) => ({
            label: u.mm_display_name,
            value: u.mm_id,
        }));

        for (const opt of options) {
            optionCache.current[opt.value] = opt.label;
        }

        return options;
    }, []);

    const isValidEmail = (input: string): boolean => {
        return (/\S+@\S+\.\S+/).test(input);
    };

    const handleChange = (selected: readonly SelectOption[] | null) => {
        if (selected) {
            for (const opt of selected) {
                optionCache.current[opt.value] = opt.label;
            }
        }
        props.onChange(selected ? selected.map((option) => option.value) : []);
    };

    const selectedValues = props.value.map((v) => ({
        label: optionCache.current[v] || v,
        value: v,
    }));

    return (
        <>
            <AsyncCreatableSelect<SelectOption, true>
                value={selectedValues}
                loadOptions={loadOptions}
                defaultOptions={true}
                menuPortalTarget={document.body}
                menuPlacement='auto'
                onChange={handleChange}
                isValidNewOption={isValidEmail}
                styles={getStyleForReactSelect(theme)}
                isMulti={true}
            />
            {storedError && (
                <div>
                    <span className='error-text'>{storedError}</span>
                </div>
            )}
        </>
    );
}
