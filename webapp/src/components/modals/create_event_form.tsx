// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, { useState } from 'react';
import { useSelector } from 'react-redux';

import { Modal } from 'react-bootstrap';

import { getTheme } from 'mattermost-redux/selectors/entities/preferences';

import { CreateEventPayload } from '@/types/calendar_api_types';

import { getModalStyles } from '@/utils/styles';

import FormButton from '@/components/form_button';
import Loading from '@/components/loading';
import Setting from '@/components/setting';
import AttendeeSelector from '@/components/attendee_selector';
import TimeSelector from '@/components/time_selector';
import { doFetchWithResponse } from '@/client';
import ChannelSelector from '../channel_selector';

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
        date: '',
        start_time: '',
        end_time: '',
        description: '',
        channel_id: '',
    });

    const setFormValue = <Key extends keyof CreateEventPayload>(name: Key, value: CreateEventPayload[Key]) => {
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
        setSubmitting(false);
    };

    const createEvent = async (payload: CreateEventPayload): Promise<{ error?: string, data?: any }> => {
        return new Promise((resolve, reject) => {
            doFetchWithResponse('/plugins/com.mattermost.gcal/api/v1/events/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            })
                .then((response) => {
                    return response.data
                })
                .then((data) => {
                    if (data && data.error) {
                        reject({ error: data.error });
                    } else {
                        resolve({ data });
                    }
                })
                .catch((error) => {
                    reject({ 'error': error });
                });
        });
    };

    const handleSubmit = (e?: React.FormEvent) => {
        if (e && e.preventDefault) {
            e.preventDefault();
        }

        // add required field validation

        setSubmitting(true);
        createEvent(formValues).then(({ data }) => {

            handleClose();
        }).catch(({ error }) => {

            if (error) {
                handleError(error);
                return;
            }
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
        form = <Loading />;
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
                    style={{ marginRight: '10px' }}
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
    setFormValue: <Key extends keyof CreateEventPayload>(name: Key, value: CreateEventPayload[Key]) => Promise<{ error?: string }>;
}

const ActualForm = (props: ActualFormProps) => {
    const { formValues, setFormValue } = props;

    const theme = useSelector(getTheme);

    const components = [
        {
            label: 'Subject',
            required: true,
            component: (
                <input
                    onChange={(e) => setFormValue('subject', e.target.value)}
                    value={formValues.subject}
                    className='form-control'
                />
            ),
        },
        {
            label: 'Guests (optional)',
            component: (
                <AttendeeSelector
                    onChange={(selected) => setFormValue('attendees', selected)}
                />
            ),
        },
        {
            label: 'Date',
            required: true,
            component: (
                <input
                    onChange={(e) => setFormValue('date', e.target.value)}
                    value={formValues.date}
                    className='form-control'
                    type='date'
                />
            ),
        },
        {
            label: 'Start Time',
            required: true,
            component: (
                <TimeSelector
                    value={formValues.start_time}
                    onChange={(value) => setFormValue('start_time', value)}
                />
            ),
        },
        {
            label: 'End Time',
            required: true,
            component: (
                <TimeSelector
                    value={formValues.end_time}
                    onChange={(value) => setFormValue('end_time', value)}
                />
            ),
        },
        {
            label: 'Description (optional)',
            component: (
                <textarea
                    onChange={(e) => setFormValue('description', e.target.value)}
                    value={formValues.description}
                    className='form-control'
                />
            ),
        },
        {
            label: 'Link event to channel',
            component: (
                <ChannelSelector
                    onChange={(selected) => setFormValue('channel_id', selected)}
                />
            ),
        },

    ];

    return (
        <div>
            {components.map((c) => (
                <Setting
                    key={c.label}
                    label={c.label}
                    inputId={c.label}
                    required={c.required}
                >
                    {c.component}
                </Setting>
            ))}
        </div>
    );
};
