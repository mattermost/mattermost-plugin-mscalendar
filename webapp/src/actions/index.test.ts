import {Channel} from '@mattermost/types/channels';
import {GlobalState} from '@mattermost/types/store';
import {haveIChannelPermission} from 'mattermost-redux/selectors/entities/roles';

import {autocompleteUserChannels, AutocompleteChannelsResponse} from './index';

const mockAutocompleteChannelsForSearch = jest.fn();
const mockAutocompleteChannels = jest.fn();

jest.mock('@mattermost/client', () => ({
    Client4: jest.fn().mockImplementation(() => ({
        setUrl: jest.fn(),
        autocompleteChannelsForSearch: (...args: unknown[]) => mockAutocompleteChannelsForSearch(...args),
        autocompleteChannels: (...args: unknown[]) => mockAutocompleteChannels(...args),
    })),
    ClientError: class extends Error {},
}));

jest.mock('mattermost-redux/selectors/entities/roles', () => ({
    haveIChannelPermission: jest.fn(),
}));

const teamId = 'team1';
const state = {entities: {general: {config: {SiteURL: 'http://localhost:8065'}}}} as unknown as GlobalState;

const channel = (id: string, displayName: string, team = teamId) => ({id, display_name: displayName, team_id: team}) as Channel;

const run = (input: string) => autocompleteUserChannels(input, teamId)(jest.fn(), () => state, null as never);

const ids = (result: AutocompleteChannelsResponse) => result.data?.map((c) => c.id);
const names = (result: AutocompleteChannelsResponse) => result.data?.map((c) => c.display_name);
const allowAllExcept = (deniedId: string) => (_state: GlobalState, _team: string, channelId: string) => channelId !== deniedId;

describe('autocompleteUserChannels', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (haveIChannelPermission as jest.Mock).mockReturnValue(true);
    });

    it('searches only channels the user is a member of', async () => {
        mockAutocompleteChannelsForSearch.mockResolvedValue([channel('c1', 'Town Square')]);

        const result = await run('town');

        expect(mockAutocompleteChannelsForSearch).toHaveBeenCalledWith(teamId, 'town');
        expect(mockAutocompleteChannels).not.toHaveBeenCalled();
        expect(ids(result)).toEqual(['c1']);
    });

    it('returns channels sorted by name when the search is empty', async () => {
        mockAutocompleteChannelsForSearch.mockResolvedValue([channel('c1', 'Town Square'), channel('c2', 'Off-Topic')]);

        const result = await run('');

        expect(mockAutocompleteChannelsForSearch).toHaveBeenCalledWith(teamId, '');
        expect(names(result)).toEqual(['Off-Topic', 'Town Square']);
    });

    it('excludes channels without post permission and group messages', async () => {
        mockAutocompleteChannelsForSearch.mockResolvedValue([
            channel('c1', 'Town Square'),
            channel('c2', 'Read Only'),
            channel('gm', 'alice, bob', ''),
        ]);
        (haveIChannelPermission as jest.Mock).mockImplementation(allowAllExcept('c2'));

        const result = await run('');

        expect(ids(result)).toEqual(['c1']);
    });

    it('returns the error message when the request fails', async () => {
        mockAutocompleteChannelsForSearch.mockRejectedValue(new Error('boom'));

        const result = await run('');

        expect(result).toEqual({data: [], error: 'boom'});
    });
});
