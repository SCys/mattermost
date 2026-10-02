// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';

import {makeGetUsersTypingDetailsByChannelAndPost} from 'mattermost-redux/selectors/entities/typing';

import type {GlobalState} from 'types/store';

import {userStartedTyping, userStoppedTyping} from './actions';
import MsgTyping from './msg_typing';

type OwnProps = {
    channelId: string;
    rootId: string;
};

function makeMapStateToProps() {
    const getUsersTypingDetailsByChannelAndPost = makeGetUsersTypingDetailsByChannelAndPost();

    return function mapStateToProps(state: GlobalState, ownProps: OwnProps) {
        const typingUsers = getUsersTypingDetailsByChannelAndPost(state, {channelId: ownProps.channelId, postId: ownProps.rootId});

        return {
            typingUsers,
        };
    };
}

const mapDispatchToProps = {
    userStartedTyping,
    userStoppedTyping,
};

export default connect(makeMapStateToProps, mapDispatchToProps)(MsgTyping);
