import React, {memo, useCallback, useEffect, useRef, useState} from 'react';
import ReactSelect, {ActionMeta, GroupBase, Props as ReactSelectBaseProps} from 'react-select';
import AsyncSelect from 'react-select/async';
import CreatableSelect from 'react-select/creatable';

import {Theme} from 'mattermost-redux/selectors/entities/preferences';

import Setting from '@/components/setting';

import {getStyleForReactSelect} from '@/utils/styles';

type ReactSelectOption = {
    label: string;
    value: string;
}

const MAX_NUM_OPTIONS = 100;

type OmitKeys<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>

export type Props = OmitKeys<ReactSelectBaseProps<ReactSelectOption, boolean, GroupBase<ReactSelectOption>>, 'theme' | 'onChange'> & {
    theme: Theme;
    label?: React.ReactNode;
    onChange?: (name: string | undefined, value: string | string[] | null) => void;
    addValidate?: (isValid: () => boolean) => void;
    removeValidate?: (isValid: () => boolean) => void;
    allowUserDefinedValue?: boolean;
    limitOptions?: boolean;
    resetInvalidOnChange?: boolean;
};

function getComparableValue(v: Props['value']): string | null {
    if (!v) {
        return null;
    }
    if (Array.isArray(v)) {
        return (v as ReactSelectOption[]).map((o) => o.value).sort().join(',');
    }
    return (v as ReactSelectOption).value;
}

function ReactSelectSetting(props: Props) {
    const {
        theme,
        label,
        addValidate,
        removeValidate,
        allowUserDefinedValue,
        limitOptions,
        resetInvalidOnChange,
        onChange,
        ...selectProps
    } = props;

    const [invalid, setInvalid] = useState(false);
    const prevValueRef = useRef(getComparableValue(props.value));
    const valueRef = useRef(props.value);
    const requiredRef = useRef(props.required);
    valueRef.current = props.value;
    requiredRef.current = props.required;

    const isValid = useCallback(() => {
        if (!requiredRef.current) {
            return true;
        }

        let valid = Boolean(valueRef.current);
        if (valueRef.current && Array.isArray(valueRef.current)) {
            valid = Boolean(valueRef.current.length);
        }

        setInvalid(!valid);
        return valid;
    }, []);

    useEffect(() => {
        addValidate?.(isValid);
        return () => {
            removeValidate?.(isValid);
        };
    }, [addValidate, removeValidate, isValid]);

    useEffect(() => {
        const current = getComparableValue(props.value);
        if (invalid && current !== prevValueRef.current) {
            setInvalid(false);
        }
        prevValueRef.current = current;
    }, [props.value, invalid]);

    const handleChange = useCallback((value: readonly ReactSelectOption[] | ReactSelectOption | null, _action: ActionMeta<ReactSelectOption>) => {
        if (onChange) {
            if (Array.isArray(value)) {
                onChange(props.name, (value as ReactSelectOption[]).map((x) => x.value));
            } else {
                const single = value as ReactSelectOption | null;
                onChange(props.name, single ? single.value : null);
            }
        }
        if (resetInvalidOnChange) {
            setInvalid(false);
        }
    }, [onChange, props.name, resetInvalidOnChange]);

    const filterOptions = useCallback((input: string) => {
        let options = props.options ?? [];
        if (input) {
            options = options.filter((opt) => 'label' in opt && opt.label?.toUpperCase().includes(input.toUpperCase()));
        }
        return Promise.resolve(options.slice(0, MAX_NUM_OPTIONS));
    }, [props.options]);

    const requiredMsg = 'This field is required.';
    let validationError = null;

    if (props.required && invalid) {
        validationError = (
            <p className='help-text error-text'>
                <span>{requiredMsg}</span>
            </p>
        );
    }

    let selectComponent = null;
    if (limitOptions && (selectProps.options?.length ?? 0) > MAX_NUM_OPTIONS) {
        selectComponent = (
            <AsyncSelect
                {...selectProps}
                loadOptions={filterOptions}
                defaultOptions={true}
                menuPortalTarget={document.body}
                menuPlacement='auto'
                onChange={handleChange}
                styles={getStyleForReactSelect(theme)}
            />
        );
    } else if (allowUserDefinedValue) {
        selectComponent = (
            <CreatableSelect
                {...selectProps}
                noOptionsMessage={() => 'Start typing...'}
                formatCreateLabel={(value) => `Add "${value}"`}
                placeholder=''
                menuPortalTarget={document.body}
                menuPlacement='auto'
                onChange={handleChange}
                styles={getStyleForReactSelect(theme)}
            />
        );
    } else {
        selectComponent = (
            <ReactSelect
                {...selectProps}
                menuPortalTarget={document.body}
                menuPlacement='auto'
                onChange={handleChange}
                styles={getStyleForReactSelect(theme)}
            />
        );
    }

    return (
        <Setting
            inputId={props.name}
            label={props.label}
            required={props.required}
        >
            {selectComponent}
            {validationError}
        </Setting>
    );
}

export default memo(ReactSelectSetting);
