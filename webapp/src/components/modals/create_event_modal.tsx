// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState} from 'react';

import {Modal} from 'react-bootstrap';

import CreateEventForm from './create_event_form';

type Props = {
    // visible: boolean;
    // close: () => void;
}

export default function CreateEventModal(props: Props) {
    const [visible, setVisible] = useState(false);
    if (!visible) {
        return null;
    }

    const close = () => {
        setVisible(false);
    }

    return (
        <Modal
            dialogClassName='modal--scroll'
            show={visible}
            onHide={close}
            onExited={close}
            bsSize='large'
            backdrop='static'
        >
            <Modal.Header closeButton={true}>
                <Modal.Title>{'Create Calendar Event'}</Modal.Title>
            </Modal.Header>
            <CreateEventForm {...props} close={close}/>
        </Modal>
    );
}
