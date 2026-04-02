import React, {PureComponent} from 'react';
import PropTypes from 'prop-types';

let formButtonCounter = 0;

export default class FormButton extends PureComponent {
    static propTypes = {
        id: PropTypes.string,
        executing: PropTypes.bool,
        disabled: PropTypes.bool,
        executingMessage: PropTypes.node,
        defaultMessage: PropTypes.node,
        btnClass: PropTypes.string,
        extraClasses: PropTypes.string,
        saving: PropTypes.bool,
        savingMessage: PropTypes.string,
        type: PropTypes.string,
    };

    static defaultProps = {
        disabled: false,
        savingMessage: 'Creating',
        defaultMessage: 'Create',
        btnClass: 'btn-primary',
        extraClasses: '',
        type: 'button',
    };

    constructor(props) {
        super(props);
        formButtonCounter += 1;
        this.generatedId = `formButton-${formButtonCounter}`;
    }

    render() {
        const {
            saving,
            disabled,
            executing,
            executingMessage,
            savingMessage,
            defaultMessage,
            btnClass,
            extraClasses,
            id,
            ...props
        } = this.props;
        const buttonId = id || this.generatedId;

        let contents;
        if (executing) {
            contents = (
                <span>
                    <span
                        className='fa fa-spin fa-spinner'
                        title={'Loading Icon'}
                        aria-hidden={true}
                    />
                    {executingMessage || savingMessage}
                </span>
            );
        } else if (saving) {
            contents = (
                <span>
                    <span
                        className='fa fa-spin fa-spinner'
                        title={'Loading Icon'}
                        aria-hidden={true}
                    />
                    {savingMessage}
                </span>
            );
        } else {
            contents = defaultMessage;
        }

        let className = 'save-button btn ' + btnClass;

        if (extraClasses) {
            className += ' ' + extraClasses;
        }

        return (
            <button
                id={buttonId}
                className={className}
                disabled={disabled || saving || executing}
                {...props}
            >
                {contents}
            </button>
        );
    }
}
