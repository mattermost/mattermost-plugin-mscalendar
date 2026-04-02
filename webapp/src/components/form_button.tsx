import React, {memo, useId} from 'react';

type Props = React.ButtonHTMLAttributes<HTMLButtonElement> & {
    id?: string;
    executing?: boolean;
    disabled?: boolean;
    executingMessage?: React.ReactNode;
    defaultMessage?: React.ReactNode;
    btnClass?: string;
    extraClasses?: string;
    saving?: boolean;
    savingMessage?: React.ReactNode;
    type?: 'button' | 'submit' | 'reset';
};

function FormButton({
    id,
    executing = false,
    disabled = false,
    executingMessage,
    defaultMessage = 'Create',
    btnClass = 'btn-primary',
    extraClasses = '',
    saving = false,
    savingMessage = 'Creating',
    type = 'button',
    children,
    ...rest
}: Props) {
    const generatedId = useId();
    const buttonId = id || generatedId;

    let contents: React.ReactNode;
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
        contents = children ?? defaultMessage;
    }

    let className = 'save-button btn ' + btnClass;
    if (extraClasses) {
        className += ' ' + extraClasses;
    }
    if (rest.className) {
        className += ' ' + rest.className;
    }

    return (
        <button
            {...rest}
            id={buttonId}
            className={className}
            disabled={disabled || saving || executing}
            aria-busy={saving || executing}
            type={type}
        >
            {contents}
        </button>
    );
}

export default memo(FormButton);
