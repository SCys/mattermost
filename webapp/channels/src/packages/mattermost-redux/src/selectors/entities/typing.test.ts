// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {
    makeGetUsersTypingByChannelAndPost,
    makeGetUsersTypingDetailsByChannelAndPost,
} from 'mattermost-redux/selectors/entities/typing';
import deepFreezeAndThrowOnMutation from 'mattermost-redux/utils/deep_freeze';

describe('Selectors.Typing', () => {
    const channelId = 'channel1';
    const postId = 'post1';
    const key = channelId + postId;

    const user1 = {
        id: 'user1',
        username: 'alice',
        first_name: 'Alice',
        last_name: 'Smith',
        nickname: '',
        is_bot: false,
    };

    const bot1 = {
        id: 'bot1',
        username: 'agent-creative',
        first_name: 'Creative',
        last_name: 'Agent',
        nickname: '',
        is_bot: true,
    };

    const testState = deepFreezeAndThrowOnMutation({
        entities: {
            users: {
                profiles: {
                    user1,
                    bot1,
                },
            },
            preferences: {
                myPreferences: {},
            },
            typing: {
                [key]: {
                    user1: 1234,
                    bot1: 1235,
                },
            },
        },
    });

    test('makeGetUsersTypingDetailsByChannelAndPost should return typing user details with isBot flag', () => {
        const getTypingDetails = makeGetUsersTypingDetailsByChannelAndPost();
        const details = getTypingDetails(testState, {channelId, postId});

        expect(details).toEqual([
            {
                id: 'user1',
                name: 'alice',
                isBot: false,
            },
            {
                id: 'bot1',
                name: 'agent-creative',
                isBot: true,
            },
        ]);
    });

    test('makeGetUsersTypingByChannelAndPost should return string names array', () => {
        const getTypingUsers = makeGetUsersTypingByChannelAndPost();
        const users = getTypingUsers(testState, {channelId, postId});

        expect(users).toEqual(['alice', 'agent-creative']);
    });

    test('should return empty array when no users are typing', () => {
        const getTypingDetails = makeGetUsersTypingDetailsByChannelAndPost();
        const details = getTypingDetails(testState, {channelId: 'otherChannel', postId: ''});

        expect(details).toEqual([]);
    });
});
