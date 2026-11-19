import {Channel, ChannelMembership} from '@mattermost/types/channels';
import {GlobalState} from '@mattermost/types/store';
import {haveIChannelPermission} from 'mattermost-redux/selectors/entities/roles';

import {autocompleteUserChannels, AutocompleteChannelsResponse} from './index';

jest.mock('mattermost-redux/selectors/entities/roles', () => ({
    haveIChannelPermission: jest.fn(),
}));

const teamId = 'team1';

const channel = (id: string, displayName: string, overrides: Partial<Channel> = {}) => ({
    id,
    name: displayName.toLowerCase().replace(/\s+/g, '-'),
    display_name: displayName,
    team_id: teamId,
    type: 'O',
    delete_at: 0,
    ...overrides,
}) as Channel;

const makeState = (channels: Channel[], channelsInTeam: Record<string, string[]>, memberOf: string[]) => ({
    entities: {
        channels: {
            channels: Object.fromEntries(channels.map((c) => [c.id, c])),
            channelsInTeam: Object.fromEntries(Object.entries(channelsInTeam).map(([team, ids]) => [team, new Set(ids)])),
            myMembers: Object.fromEntries(memberOf.map((id) => [id, {channel_id: id} as ChannelMembership])),
        },
    },
}) as unknown as GlobalState;

const run = (state: GlobalState, input: string) => autocompleteUserChannels(input, teamId)(jest.fn(), () => state, null as never);

const ids = (result: AutocompleteChannelsResponse) => result.data?.map((c) => c.id);
const unjoinedChannels = (count: number) => Array.from({length: count}, (_, i) => channel(`u${i}`, `A unjoined ${i}`));
const channelIds = (channels: Channel[]) => channels.map((c) => c.id);
const allowAllExcept = (deniedId: string) => (_state: GlobalState, _team: string, channelId: string) => channelId !== deniedId;

describe('autocompleteUserChannels', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (haveIChannelPermission as jest.Mock).mockReturnValue(true);
    });

    it('returns only channels the user is a member of, sorted by name', async () => {
        const state = makeState(
            [channel('c1', 'Town Square'), channel('c2', 'Off-Topic'), channel('c3', 'Unjoined')],
            {[teamId]: ['c1', 'c2', 'c3']},
            ['c1', 'c2'],
        );

        expect(ids(await run(state, ''))).toEqual(['c2', 'c1']);
    });

    it('is not limited to the first 50 channels in the team', async () => {
        const unjoined = unjoinedChannels(60);
        const mine = channel('mine', 'Zebra');
        const state = makeState([...unjoined, mine], {[teamId]: [...channelIds(unjoined), 'mine']}, ['mine']);

        expect(ids(await run(state, ''))).toEqual(['mine']);
    });

    it('filters by display name or name, ignoring case', async () => {
        const state = makeState(
            [channel('c1', 'Town Square'), channel('c2', 'Dev Team', {name: 'engineering'})],
            {[teamId]: ['c1', 'c2']},
            ['c1', 'c2'],
        );

        expect(ids(await run(state, 'TOWN'))).toEqual(['c1']);
        expect(ids(await run(state, 'engin'))).toEqual(['c2']);
    });

    it('excludes archived channels, channels in other teams and channels without post permission', async () => {
        const state = makeState(
            [
                channel('c1', 'Town Square'),
                channel('archived', 'Archived', {delete_at: 1}),
                channel('readonly', 'Read Only'),
                channel('other', 'Other Team', {team_id: 'team2'}),
            ],
            {[teamId]: ['c1', 'archived', 'readonly'], team2: ['other']},
            ['c1', 'archived', 'readonly', 'other'],
        );
        (haveIChannelPermission as jest.Mock).mockImplementation(allowAllExcept('readonly'));

        expect(ids(await run(state, ''))).toEqual(['c1']);
    });

    it('returns no channels when the team has not been loaded', async () => {
        const state = makeState([], {}, []);

        expect(ids(await run(state, ''))).toEqual([]);
    });
});
