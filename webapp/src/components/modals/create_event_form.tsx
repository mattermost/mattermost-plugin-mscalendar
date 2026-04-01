import React, {useState} from 'react';
import {useSelector} from 'react-redux';

import {Modal as BootstrapModal} from 'react-bootstrap';

type ModalSectionProps = React.PropsWithChildren<{ style?: React.CSSProperties }>;
const ModalBody = BootstrapModal.Body as unknown as React.FC<ModalSectionProps>;
const ModalFooter = BootstrapModal.Footer as unknown as React.FC<ModalSectionProps>;

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import ChannelSelector from '../channel_selector';

import {CreateEventPayload} from '@/types/calendar_api_types';

import {getModalStyles} from '@/utils/styles';

import FormButton from '@/components/form_button';
import Loading from '@/components/loading';
import Setting from '@/components/setting';
import AttendeeSelector from '@/components/attendee_selector';
import TimeSelector from '@/components/time_selector';
import DateInput from '@/components/date_input';
import {capitalizeFirstCharacter} from '@/utils/text';
import {useAppDispatch} from '@/hooks';
import {createCalendarEvent, refreshActiveCalendarView} from '@/actions';
import {getTodayString} from '@/utils/datetime';
import {getCreateEventModal} from '@/selectors';

import './create_event_form.scss';

type Props = {
    close: (e?: Event) => void;
};

export default function CreateEventForm(props: Props) {
    const [storedError, setStoredError] = useState('');
    const [submitting, setSubmitting] = useState(false);
    const [loading, setLoading] = useState(false);

    const dispatch = useAppDispatch();
    const modalData = useSelector(getCreateEventModal);

    const [formValues, setFormValues] = useState<CreateEventPayload>({
        subject: '',
        all_day: false,
        attendees: [],
        date: modalData?.date || getTodayString(),
        start_time: modalData?.startTime || '',
        end_time: modalData?.endTime || '',
        description: '',
        channel_id: modalData?.channelId || '',
        location: '',
    });

    const setFormValue = <Key extends keyof CreateEventPayload>(name: Key, value: CreateEventPayload[Key]) => {
        setFormValues((values: CreateEventPayload) => ({
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
        const errorMessage = capitalizeFirstCharacter(error);
        setStoredError(errorMessage);
        setSubmitting(false);
    };

    const handleSubmit = async (e?: React.FormEvent) => {
        if (e && e.preventDefault) {
            e.preventDefault();
        }

        setSubmitting(true);

        const response = await dispatch(createCalendarEvent(formValues));
        if (response.error) {
            handleError(response.error);
            return;
        }

        await dispatch(refreshActiveCalendarView());
        handleClose();
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
            <ModalBody
                style={style.modalBody}
            >
                {error}
                {form}
            </ModalBody>
            <ModalFooter style={style.modalFooter}>
                {footer}
            </ModalFooter>
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
            label: 'Location (optional)',
            required: false,
            component: (
                <input
                    onChange={(e) => setFormValue('location', e.target.value)}
                    value={formValues.location}
                    className='form-control'
                />
            ),
        },
        {
            label: 'Guests (optional)',
            component: (
                <AttendeeSelector
                    value={formValues.attendees}
                    onChange={(selected) => setFormValue('attendees', selected)}
                />
            ),
        },
        {
            label: 'Date',
            required: true,
            component: (
                <DateInput
                    value={formValues.date}
                    min={getTodayString()}
                    onChange={(value) => {
                        setFormValue('date', value);
                        setFormValue('start_time', '');
                        setFormValue('end_time', '');
                    }}
                    className='form-control'
                />
            ),
        },
        {
            label: 'Start Time',
            required: true,
            component: (
                <TimeSelector
                    value={formValues.start_time}
                    endTime={formValues.end_time}
                    date={formValues.date}
                    onChange={(name: keyof CreateEventPayload, value: string) => setFormValue(name, value)}
                />
            ),
        },
        {
            label: 'End Time',
            required: true,
            component: (
                <TimeSelector
                    value={formValues.end_time}
                    startTime={formValues.start_time}
                    date={formValues.date}
                    onChange={(name: keyof CreateEventPayload, value: string) => setFormValue(name, value)}
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
            label: 'Link event to channel (optional)',
            component: (
                <ChannelSelector
                    value={formValues.channel_id ? [formValues.channel_id] : []}
                    onChange={(selected) => setFormValue('channel_id', selected)}
                />
            ),
        },

    ];

    return (
        <div className='mscalendar-create-event-form'>
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
