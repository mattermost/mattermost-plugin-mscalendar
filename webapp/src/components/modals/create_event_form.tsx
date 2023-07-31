// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState} from 'react';

// import {Modal} from 'react-bootstrap';
import {Modal} from 'react-bootstrap';

// import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

const getTheme = state => {
    return state.entities.preferences.myPreferences['theme--'];
}

import {getModalStyles} from '@/utils/styles';

import FormButton from '@/components/form_button';
import Loading from '@/components/loading';
import {useSelector} from 'react-redux';
import {CreateEventPayload} from '@/types/calendar_api_types';
// import ReactSelectSetting from '@/components/react_select_setting';

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
        setFormValues(values => ({
            ...values,
            [name]: value,
        }));
    }

    const theme = useSelector(getTheme);

    const handleClose = (e?: Event) => {
        if (e && e.preventDefault) {
            e.preventDefault();
        }

        props.close();
    }

    const handleError = (error: string) => {
        setStoredError(error);
    }

    const createEvent = async (payload: CreateEventPayload) => {

    }

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
    }

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
        form = <ActualForm formValues={formValues} setFormValue={setFormValue}/>;
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

    return (
        <div>
            <div>
                <label>
                    {'Subject:'}
                </label>
                <input
                    onChange={(e) => setFormValue('subject', e.target.value)}
                    value={formValues.subject}
                    className='form-control'
                />
            </div>
            <div>
                <label>
                    {'Start Time:'}
                </label>
                <input
                    onChange={(e) => setFormValue('start_time', e.target.value)}
                    value={formValues.start_time}
                    className='form-control'
                />
            </div>
            <div>
                <label>
                    {'End Time:'}
                </label>
                <input
                    onChange={(e) => setFormValue('end_time', e.target.value)}
                    value={formValues.end_time}
                    className='form-control'
                />
            </div>
        </div>
    )
};
