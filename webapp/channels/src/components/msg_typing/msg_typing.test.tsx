// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import {renderWithContext} from 'tests/react_testing_utils';

import MsgTyping from './msg_typing';

describe('components/MsgTyping', () => {
    const baseProps = {
        typingUsers: [],
        channelId: 'test',
        rootId: '',
        userStartedTyping: jest.fn(),
        userStoppedTyping: jest.fn(),
    };

    test('should match snapshot, on nobody typing', () => {
        const {container} = renderWithContext(<MsgTyping {...baseProps}/>);
        expect(container).toMatchSnapshot();
    });

    test('should match snapshot, on one user typing', () => {
        const typingUsers = ['test.user'];
        const props = {...baseProps, typingUsers};

        const {container} = renderWithContext(<MsgTyping {...props}/>);
        expect(container).toMatchSnapshot();
    });

    test('should match snapshot, on multiple users typing', () => {
        const typingUsers = ['test.user', 'other.test.user', 'another.user'];
        const props = {...baseProps, typingUsers};

        const {container} = renderWithContext(<MsgTyping {...props}/>);
        expect(container).toMatchSnapshot();
    });

    test('should render bot is working when one bot is typing', () => {
        const typingUsers = [{id: 'bot.user.id', name: 'agent-creative', isBot: true}];
        const props = {...baseProps, typingUsers};

        const {container} = renderWithContext(<MsgTyping {...props}/>);
        expect(container).toHaveTextContent('🤖 agent-creative is working...');
        expect(container.querySelector('.msg-typing--bot')).toBeInTheDocument();
    });

    test('should render bots are working when multiple bots are typing', () => {
        const typingUsers = [
            {id: 'bot.user.1', name: 'agent-creative', isBot: true},
            {id: 'bot.user.2', name: 'agent-helper', isBot: true},
        ];
        const props = {...baseProps, typingUsers};

        const {container} = renderWithContext(<MsgTyping {...props}/>);
        expect(container).toHaveTextContent('🤖 agent-creative and agent-helper are working...');
        expect(container.querySelector('.msg-typing--bot')).toBeInTheDocument();
    });

    test('should render standard typing message when bot and human are typing together', () => {
        const typingUsers = [
            {id: 'human.user', name: 'alice', isBot: false},
            {id: 'bot.user', name: 'agent-creative', isBot: true},
        ];
        const props = {...baseProps, typingUsers};

        const {container} = renderWithContext(<MsgTyping {...props}/>);
        expect(container).toHaveTextContent('alice and agent-creative are typing...');
        expect(container.querySelector('.msg-typing--bot')).toBeInTheDocument();
    });
});
