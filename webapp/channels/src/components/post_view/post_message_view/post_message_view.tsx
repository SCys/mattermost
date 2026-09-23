// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {FormattedMessage} from 'react-intl';

import type {Post} from '@mattermost/types/posts';

import {Posts} from 'mattermost-redux/constants';
import type {Theme} from 'mattermost-redux/selectors/entities/preferences';
import {isPostEphemeral} from 'mattermost-redux/utils/post_utils';

import PostMarkdown from 'components/post_markdown';
import ShowMore from 'components/post_view/show_more';
import type {AttachmentTextOverflowType} from 'components/post_view/show_more/show_more';

import Pluggable from 'plugins/pluggable';
import PluggableErrorBoundary from 'plugins/pluggable/error_boundary';
import {PostTypes} from 'utils/constants';
import {getPostTranslatedMessage, getPostTranslation} from 'utils/post_utils';
import type {TextFormattingOptions} from 'utils/text_formatting';
import * as Utils from 'utils/utils';

import type {PostPluginComponent} from 'types/store/plugins';

import MessageBodyFooterMountNotify from './message_body_footer_mount_notify';

// These posts types must not be rendered with the collapsible "Show More" container.
const FULL_HEIGHT_POST_TYPES = new Set([
    PostTypes.CUSTOM_DATA_SPILLAGE_REPORT,
]);

type Props = {
    post: Post; /* The post to render the message for */
    enableFormatting?: boolean; /* Set to enable Markdown formatting */
    enableRichCommandUI?: boolean; /* Switch to control whether to render Discord-style command pills in message */
    options?: TextFormattingOptions; /* Options specific to text formatting */
    compactDisplay?: boolean; /* Set to render post body compactly */
    isRHS?: boolean; /* Flags if the post_message_view is for the RHS (Reply). */
    theme: Theme; /* Logged in user's theme */
    pluginPostTypes?: {
        [postType: string]: PostPluginComponent;
    }; /* Post type components from plugins */
    currentRelativeTeamUrl: string;
    overflowType?: AttachmentTextOverflowType;
    maxHeight?: number; /* The max height used by the show more component */
    showPostEditedIndicator?: boolean; /* Whether or not to render the post edited indicator */
    isChannelAutotranslated: boolean;
    userLanguage: string;

    /** Permalink previews and similar read-only surfaces. */
    disableInteractions?: boolean;

    /** Rendered inside ShowMore with the post message (e.g. read-only interactive blocks). */
    messageBodyFooter?: React.ReactNode;
};

type State = {
    collapse: boolean;
    hasOverflow: boolean;
    checkOverflow: number;
};

export default class PostMessageView extends React.PureComponent<Props, State> {
    private imageProps: any;

    static defaultProps = {
        options: {},
        isRHS: false,
        pluginPostTypes: {},
        overflowType: undefined,
    };

    constructor(props: Props) {
        super(props);

        this.state = {
            collapse: true,
            hasOverflow: false,
            checkOverflow: 0,
        };

        this.imageProps = {
            onImageLoaded: this.handleHeightReceived,
            onImageHeightChanged: this.checkPostOverflow,
        };
    }

    checkPostOverflow = () => {
        // Increment checkOverflow to indicate change in height
        // and recompute textContainer height at ShowMore component
        // and see whether overflow text of show more/less is necessary or not.
        this.setState((prevState) => {
            return {checkOverflow: prevState.checkOverflow + 1};
        });
    };

    handleHeightReceived = (height: number) => {
        if (height > 0) {
            this.checkPostOverflow();
        }
    };

    renderDeletedPost() {
        return (
            <p>
                <FormattedMessage
                    id='post_body.deleted'
                    defaultMessage='(message deleted)'
                />
            </p>
        );
    }

    handleFormattedTextClick = (e: React.MouseEvent<HTMLDivElement, MouseEvent>) =>
        Utils.handleFormattedTextClick(e, this.props.currentRelativeTeamUrl);

    render() {
        const {
            post,
            enableFormatting,
            options,
            pluginPostTypes,
            compactDisplay,
            isRHS,
            theme,
            overflowType,
            maxHeight,
            disableInteractions,
            messageBodyFooter,
        } = this.props;

        if (post.state === Posts.POST_DELETED) {
            return this.renderDeletedPost();
        }

        if (!enableFormatting) {
            return <span>{post.message}</span>;
        }

        const postType = typeof post.props?.type === 'string' ? post.props.type : post.type;

        if (pluginPostTypes && Object.hasOwn(pluginPostTypes, postType)) {
            const PluginComponent = pluginPostTypes[postType].component;
            return (
                <PluggableErrorBoundary pluginId={pluginPostTypes[postType].pluginId}>
                    <PluginComponent
                        post={post}
                        compactDisplay={compactDisplay}
                        isRHS={isRHS}
                        theme={theme}
                    />
                </PluggableErrorBoundary>
            );
        }

        let message = post.message;
        const isEphemeral = isPostEphemeral(post);
        if (compactDisplay && isEphemeral) {
            const visibleMessage = Utils.localizeMessage({id: 'post_info.message.visible.compact', defaultMessage: ' (Only visible to you)'});
            message = message.concat(visibleMessage);
        }

        // Use translation if channel is autotranslated and translation is available
        const translation = getPostTranslation(post, this.props.userLanguage);
        if (this.props.isChannelAutotranslated && post.type === '' && translation?.state === 'ready') {
            message = getPostTranslatedMessage(message, translation);
        }

        const id = isRHS ? `rhsPostMessageText_${post.id}` : `postMessageText_${post.id}`;

        const renderCommandPill = () => {
            if (!this.props.enableRichCommandUI || !message || !message.startsWith('/')) {
                return null;
            }
            const match = message.match(/^\/([a-zA-Z0-9_-]+)(?:\s+(.*))?$/s);
            if (!match) {
                return null;
            }
            const trigger = match[1];
            const rest = (match[2] || '').trim();
            const slots: Array<{name: string; value: string}> = [];
            if (rest) {
                const flagMatches = [...rest.matchAll(/--([a-zA-Z0-9_-]+)(?:\s+([^\s-]+(?:[^-][^\s-]*)*))?/g)];
                if (flagMatches.length > 0) {
                    flagMatches.forEach((m) => {
                        slots.push({name: m[1], value: m[2] ? m[2].trim() : 'true'});
                    });
                } else {
                    slots.push({name: 'args', value: rest});
                }
            }

            return (
                <div
                    className='rich-command-post-pill'
                    style={{
                        display: 'inline-flex',
                        flexWrap: 'wrap',
                        alignItems: 'center',
                        gap: '6px',
                        padding: '4px 8px',
                        borderRadius: '6px',
                        background: 'rgba(var(--center-channel-color-rgb, 61, 60, 64), 0.05)',
                        border: '1px solid rgba(var(--center-channel-color-rgb, 61, 60, 64), 0.12)',
                        margin: '2px 0 6px 0',
                        fontSize: '13px',
                        userSelect: 'none',
                    }}
                >
                    <span
                        style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '4px',
                            fontWeight: 600,
                            color: 'var(--button-bg, #1c58d9)',
                            background: 'rgba(var(--button-bg-rgb, 28, 88, 217), 0.12)',
                            padding: '2px 6px',
                            borderRadius: '4px',
                        }}
                    >
                        <span>/</span>
                        <span>{trigger}</span>
                    </span>
                    {slots.map((s, idx) => (
                        <span
                            key={idx}
                            style={{
                                display: 'inline-flex',
                                alignItems: 'center',
                                gap: '4px',
                                background: 'rgba(var(--center-channel-color-rgb, 61, 60, 64), 0.08)',
                                padding: '2px 6px',
                                borderRadius: '4px',
                                color: 'var(--center-channel-color, #3d3c40)',
                            }}
                        >
                            <span style={{opacity: 0.7, fontWeight: 500}}>{s.name === 'args' ? '' : `${s.name}:`}</span>
                            <span style={{fontWeight: 400}}>{s.value}</span>
                        </span>
                    ))}
                </div>
            );
        };

        const commandPill = renderCommandPill();

        const body = (
            <>
                <div
                    id={id}
                    className='post-message__text'
                    data-testid='post-message-text'
                    dir='auto'
                    onClick={this.handleFormattedTextClick}
                >
                    {commandPill ? (
                        commandPill
                    ) : (
                        <PostMarkdown
                            message={message}
                            imageProps={this.imageProps}
                            options={options}
                            post={post}
                            channelId={post.channel_id}
                            showPostEditedIndicator={this.props.showPostEditedIndicator}
                            isRHS={isRHS}
                            disableInteractions={disableInteractions}
                        />
                    )}
                </div>
                {messageBodyFooter != null && (
                    <MessageBodyFooterMountNotify onHeightChange={this.checkPostOverflow}>
                        {messageBodyFooter}
                    </MessageBodyFooterMountNotify>
                )}
                <Pluggable
                    pluggableName='PostMessageAttachment'
                    postId={post.id}
                    onHeightChange={this.handleHeightReceived}
                />
            </>
        );

        if (FULL_HEIGHT_POST_TYPES.has(postType)) {
            return body;
        }

        return (
            <ShowMore
                checkOverflow={this.state.checkOverflow}
                text={message}
                overflowType={overflowType}
                maxHeight={maxHeight}
            >
                {body}
            </ShowMore>
        );
    }
}
