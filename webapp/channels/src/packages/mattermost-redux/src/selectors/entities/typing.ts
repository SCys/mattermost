// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {GlobalState} from '@mattermost/types/store';
import type {Typing} from '@mattermost/types/typing';
import type {UserProfile} from '@mattermost/types/users';
import type {IDMappedObjects} from '@mattermost/types/utilities';

import {createSelector} from 'mattermost-redux/selectors/create_selector';
import {getUsers} from 'mattermost-redux/selectors/entities/common';
import {getTeammateNameDisplaySetting} from 'mattermost-redux/selectors/entities/preferences';
import {displayUsername} from 'mattermost-redux/utils/user_utils';

export type TypingUser = {
    id: string;
    name: string;
    isBot: boolean;
};

const getUsersTypingDetailsImpl = (
    profiles: IDMappedObjects<UserProfile>,
    teammateNameDisplay: string,
    channelId: string,
    parentPostId: string,
    typing: Typing,
): TypingUser[] => {
    const id = channelId + parentPostId;

    if (typing[id]) {
        const users = Object.keys(typing[id]);

        if (users.length) {
            return users.map((userId) => {
                const profile = profiles[userId];
                return {
                    id: userId,
                    name: displayUsername(profile, teammateNameDisplay),
                    isBot: Boolean(profile?.is_bot),
                };
            });
        }
    }

    return [];
};

export function makeGetUsersTypingDetailsByChannelAndPost(): (
    state: GlobalState,
    props: {channelId: string; postId: string},
) => TypingUser[] {
    return createSelector(
        'makeGetUsersTypingDetailsByChannelAndPost',
        getUsers,
        getTeammateNameDisplaySetting,
        (state: GlobalState, options: {channelId: string; postId: string}) => options.channelId,
        (state: GlobalState, options: {channelId: string; postId: string}) => options.postId,
        (state: GlobalState) => state.entities.typing,
        getUsersTypingDetailsImpl,
    );
}

export function makeGetUsersTypingByChannelAndPost(): (state: GlobalState, props: {channelId: string; postId: string}) => string[] {
    const getUsersTypingDetails = makeGetUsersTypingDetailsByChannelAndPost();
    return createSelector(
        'makeGetUsersTypingByChannelAndPost',
        (state: GlobalState, options: {channelId: string; postId: string}) => getUsersTypingDetails(state, options),
        (users) => users.map((user) => user.name),
    );
}
