// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useCallback} from 'react';
import {FormattedMessage} from 'react-intl';

import {WebSocketEvents, type WebSocketMessage} from '@mattermost/client';

import {useWebSocket} from 'utils/use_websocket';

export type TypingUserInfo = {
    id?: string;
    name: string;
    isBot?: boolean;
};

type Props = {
    channelId: string;
    rootId: string;
    typingUsers: Array<string | TypingUserInfo>;
    userStartedTyping: (userId: string, channelId: string, rootId: string, now: number) => void;
    userStoppedTyping: (userId: string, channelId: string, rootId: string, now: number) => void;
};

export default function MsgTyping(props: Props) {
    const {userStartedTyping, userStoppedTyping} = props;
    useWebSocket({
        handler: useCallback((msg: WebSocketMessage) => {
            if (msg.event === WebSocketEvents.Typing) {
                const channelId = msg.broadcast.channel_id;
                const rootId = msg.data.parent_id;
                const userId = msg.data.user_id;

                if (props.channelId === channelId && props.rootId === rootId) {
                    userStartedTyping(userId, channelId, rootId, Date.now());
                }
            } else if (msg.event === WebSocketEvents.Posted) {
                const post = JSON.parse(msg.data.post);

                const channelId = post.channel_id;
                const rootId = post.root_id;
                const userId = post.user_id;

                if (props.channelId === channelId && props.rootId === rootId) {
                    userStoppedTyping(userId, channelId, rootId, Date.now());
                }
            }
        }, [props.channelId, props.rootId, userStartedTyping, userStoppedTyping]),
    });

    const getTypingText = () => {
        const users: TypingUserInfo[] = (props.typingUsers || []).map((u) => {
            if (typeof u === 'string') {
                return {name: u, isBot: false};
            }
            return u;
        });

        const numUsers = users.length;
        if (numUsers === 0) {
            return '';
        }

        if (numUsers === 1) {
            const user = users[0];
            if (user.isBot) {
                return (
                    <FormattedMessage
                        id='msg_typing.isWorking'
                        defaultMessage='🤖 {user} is working...'
                        values={{
                            user: user.name,
                        }}
                    />
                );
            }

            return (
                <FormattedMessage
                    id='msg_typing.isTyping'
                    defaultMessage='{user} is typing...'
                    values={{
                        user: user.name,
                    }}
                />
            );
        }

        const allBots = users.every((u) => u.isBot);
        const names = users.map((u) => u.name);
        const last = names.pop();

        if (allBots) {
            return (
                <FormattedMessage
                    id='msg_typing.areWorking'
                    defaultMessage='🤖 {users} and {last} are working...'
                    values={{
                        users: names.join(', '),
                        last,
                    }}
                />
            );
        }

        return (
            <FormattedMessage
                id='msg_typing.areTyping'
                defaultMessage='{users} and {last} are typing...'
                values={{
                    users: names.join(', '),
                    last,
                }}
            />
        );
    };

    const hasBot = (props.typingUsers || []).some((u) => typeof u !== 'string' && u.isBot);

    return (
        <span className={`msg-typing${hasBot ? ' msg-typing--bot' : ''}`}>{getTypingText()}</span>
    );
}
