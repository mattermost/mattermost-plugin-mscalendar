import React, {memo, useRef} from 'react';

let formButtonCounter = 0;

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
    const generatedIdRef = useRef<string>('');
    if (!generatedIdRef.current) {
        formButtonCounter += 1;
        generatedIdRef.current = `formButton-${formButtonCounter}`;
    }
    const buttonId = id || generatedIdRef.current;

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
        contents = children || defaultMessage;
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
            type={type}
            {...rest}
        >
            {contents}
        </button>
    );
}

export default memo(FormButton);
