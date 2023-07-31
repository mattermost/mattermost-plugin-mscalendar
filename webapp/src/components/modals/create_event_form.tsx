// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState} from 'react';
import {useSelector} from 'react-redux';

import {Modal} from 'react-bootstrap';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import {CreateEventPayload} from '@/types/calendar_api_types';

import {getModalStyles} from '@/utils/styles';

import FormButton from '@/components/form_button';
import Loading from '@/components/loading';
import ReactSelectSetting from '@/components/react_select_setting';

type Props = {
    close: (e?: Event) => void;
};

export default function CreateEventForm(props: Props) {
    const [storedError, setStoredError] = useState('');
    const [submitting, setSubmitting] = useState(false);
    const [loading, setLoading] = useState(false);

    const [formValues, setFormValues] = useState<CreateEventPayload>({
        subject: '',
        all_day: false,
        attendees: [],
        start_time: '',
        end_time: '',
        body: '',
    });

    const setFormValue = (name, value) => {
        setFormValues((values) => ({
            ...values,
            [name]: value,
        }));
    };

    const theme = useSelector(getTheme);

    const handleClose = (e?: Event) => {
        if (e && e.preventDefault) {
            e.preventDefault();
        }

        props.close();
    };

    const handleError = (error: string) => {
        setStoredError(error);
    };

    const createEvent = async (payload: CreateEventPayload) => {

    };

    const handleSubmit = (e?: React.FormEvent) => {
        if (e && e.preventDefault) {
            e.preventDefault();
        }

        setSubmitting(true);
        createEvent(event).then(({error}) => {
            if (error) {
                setStoredError(error.message);
                setSubmitting(false);
                return;
            }

            handleClose();
        });
    };

    const style = getModalStyles(theme);

    const disableSubmit = false;
    const footer = (
        <React.Fragment>
            <FormButton
                type='button'
                btnClass='btn-link'
                defaultMessage='Cancel'
                onClick={handleClose}
            />
            <FormButton
                id='submit-button'
                type='submit'
                btnClass='btn btn-primary'
                saving={submitting}
                disabled={disableSubmit}
            >
                {'Create'}
            </FormButton>
        </React.Fragment>
    );

    let form;
    if (loading) {
        form = <Loading/>;
    } else {
        form = (
            <ActualForm
                formValues={formValues}
                setFormValue={setFormValue}
            />
        );
    }

    let error;
    if (storedError) {
        error = (
            <p className='alert alert-danger'>
                <i
                    style={{marginRight: '10px'}}
                    className='fa fa-warning'
                    title='Warning Icon'
                />
                <span>{storedError}</span>
            </p>
        );
    }

    return (
        <form
            role='form'
            onSubmit={handleSubmit}
        >
            <Modal.Body
                style={style.modalBody}
            >
                {error}
                {form}
            </Modal.Body>
            <Modal.Footer style={style.modalFooter}>
                {footer}
            </Modal.Footer>
        </form>
    );
}

type ActualFormProps = {
    formValues: CreateEventPayload;
    setFormValue: <Key extends keyof CreateEventPayload>(name: Key, value: CreateEventPayload[Key]) => void;
}

const ActualForm = (props: ActualFormProps) => {
    const {formValues, setFormValue} = props;

    const theme = useSelector(getTheme);

    const attendeeOptions = [
        {label: 'sysadmin', value: 'sysadmin'},
    ];

    const attendeesSelect = (
        <ReactSelectSetting
            theme={theme}
            options={attendeeOptions}
            value={formValues.attendees}
            onChange={(selected) => setFormValue('attendees', selected)}
            isMulti={true}
        />
    );

    const components = [
        {
            label: 'Subject',
            component: (
                <input
                    onChange={(e) => setFormValue('subject', e.target.value)}
                    value={formValues.subject}
                    className='form-control'
                />
            ),
        },
        {
            label: 'Start Time',
            component: (
                <input
                    onChange={(e) => setFormValue('start_time', e.target.value)}
                    value={formValues.start_time}
                    className='form-control'
                />
            ),
        },
        {
            label: 'End Time',
            component: (
                <input
                    onChange={(e) => setFormValue('end_time', e.target.value)}
                    value={formValues.end_time}
                    className='form-control'
                />
            ),
        },
        {
            label: 'Guests',
            component: attendeesSelect,
        },
    ];

    return (
        <div>
            {components.map((c) => (
                <div
                    key={c.label}
                    className='form-group'
                >
                    <label>{c.label}</label>
                    {c.component}
                </div>
            ))}
        </div>
    );
};