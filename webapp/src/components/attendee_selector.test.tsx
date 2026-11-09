import React from 'react';
import {act, fireEvent, render, screen} from '@testing-library/react';
import '@testing-library/jest-dom';

import {mockTheme} from '@/testutils/theme';

import AttendeeSelector from './attendee_selector';

const mockDispatch = jest.fn();

jest.mock('react-redux', () => ({
    useSelector: () => mockTheme,
}));

jest.mock('@/hooks', () => ({
    useAppDispatch: () => mockDispatch,
}));

jest.mock('@/actions', () => ({
    autocompleteConnectedUsers: (input: string) => ({type: 'autocomplete', input}),
}));

const renderSelector = async () => {
    await act(async () => {
        render(
            <AttendeeSelector
                inputId='attendees'
                value={[]}
                onChange={jest.fn()}
            />,
        );
    });
};

const advanceTimers = async (ms: number) => {
    await act(async () => {
        jest.advanceTimersByTime(ms);
    });
};

describe('AttendeeSelector', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        mockDispatch.mockResolvedValue({data: []});
    });

    afterEach(() => {
        jest.useRealTimers();
        jest.clearAllMocks();
    });

    it('loads default options immediately', async () => {
        await renderSelector();

        expect(mockDispatch).toHaveBeenCalledTimes(1);
        expect(mockDispatch).toHaveBeenCalledWith({type: 'autocomplete', input: ''});
    });

    it('sends a single request for the final input when typing quickly', async () => {
        await renderSelector();
        mockDispatch.mockClear();

        const input = screen.getByRole('combobox');
        fireEvent.change(input, {target: {value: 'a'}});
        fireEvent.change(input, {target: {value: 'al'}});
        fireEvent.change(input, {target: {value: 'ali'}});

        expect(mockDispatch).not.toHaveBeenCalled();

        await advanceTimers(250);

        expect(mockDispatch).toHaveBeenCalledTimes(1);
        expect(mockDispatch).toHaveBeenCalledWith({type: 'autocomplete', input: 'ali'});
    });
});
