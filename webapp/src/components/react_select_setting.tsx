import React from 'react';
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

type State = {
    invalid: boolean;
};

export default class ReactSelectSetting extends React.PureComponent<Props, State> {
    state: State = {invalid: false};

    componentDidMount() {
        if (this.props.addValidate) {
            this.props.addValidate(this.isValid);
        }
    }

    componentWillUnmount() {
        if (this.props.removeValidate) {
            this.props.removeValidate(this.isValid);
        }
    }

    componentDidUpdate(prevProps: Props, prevState: State) {
        const getComparableValue = (v: Props['value']): string | null => {
            if (!v) {
                return null;
            }
            if (Array.isArray(v)) {
                return (v as ReactSelectOption[]).map((o) => o.value).sort().join(',');
            }
            return (v as ReactSelectOption).value;
        };
        if (prevState.invalid && getComparableValue(this.props.value) !== getComparableValue(prevProps.value)) {
            this.setState({invalid: false}); //eslint-disable-line react/no-did-update-set-state
        }
    }

    handleChange = (value: readonly ReactSelectOption[] | ReactSelectOption | null, action: ActionMeta<ReactSelectOption>) => {
        if (this.props.onChange) {
            if (Array.isArray(value)) {
                this.props.onChange(this.props.name, (value as ReactSelectOption[]).map((x) => x.value));
            } else {
                const single = value as ReactSelectOption | null;
                this.props.onChange(this.props.name, single ? single.value : null);
            }
        }
        if (this.props.resetInvalidOnChange) {
            this.setState({invalid: false});
        }
    };

    filterOptions = (input: string) => {
        let options = this.props.options ?? [];
        if (input) {
            options = options.filter((opt) => 'label' in opt && opt.label?.toUpperCase().includes(input.toUpperCase()));
        }
        return Promise.resolve(options.slice(0, MAX_NUM_OPTIONS));
    };

    isValid = () => {
        if (!this.props.required) {
            return true;
        }

        let valid = Boolean(this.props.value);
        if (this.props.value && Array.isArray(this.props.value)) {
            valid = Boolean(this.props.value.length);
        }

        this.setState({invalid: !valid});
        return valid;
    };

    render() {
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
        } = this.props;

        const requiredMsg = 'This field is required.';
        let validationError = null;

        if (this.props.required && this.state.invalid) {
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
                    loadOptions={this.filterOptions}
                    defaultOptions={true}
                    menuPortalTarget={document.body}
                    menuPlacement='auto'
                    onChange={this.handleChange}
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
                    onChange={this.handleChange}
                    styles={getStyleForReactSelect(theme)}
                />
            );
        } else {
            selectComponent = (
                <ReactSelect
                    {...selectProps}
                    menuPortalTarget={document.body}
                    menuPlacement='auto'
                    onChange={this.handleChange}
                    styles={getStyleForReactSelect(theme)}
                />
            );
        }
        return (
            <Setting
                inputId={this.props.name}
                label={this.props.label}
                required={this.props.required}
            >
                {selectComponent}
                {validationError}
            </Setting>
        );
    }
}
