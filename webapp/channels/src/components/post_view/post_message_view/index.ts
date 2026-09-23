// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';

import {Preferences} from 'mattermost-redux/constants';
import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {getTheme, getBool} from 'mattermost-redux/selectors/entities/preferences';
import {getCurrentRelativeTeamUrl} from 'mattermost-redux/selectors/entities/teams';

import type {GlobalState} from 'types/store';

import PostMessageView from './post_message_view';

function mapStateToProps(state: GlobalState) {
    const config = getConfig(state);
    const enableRichCommandUI = config.EnableRichCommandUI !== 'false';

    return {
        enableFormatting: getBool(state, Preferences.CATEGORY_ADVANCED_SETTINGS, 'formatting', true),
        enableRichCommandUI,
        pluginPostTypes: state.plugins.postTypes,
        theme: getTheme(state),
        currentRelativeTeamUrl: getCurrentRelativeTeamUrl(state),
    };
}

export default connect(mapStateToProps)(PostMessageView);
