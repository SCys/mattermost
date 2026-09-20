// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"strings"

	"github.com/mattermost/mattermost/server/public/shared/i18n"
)

type CommandOptionArg struct {
	Name     string              `json:"name"`
	Type     AutocompleteArgType `json:"type"`
	Value    any                 `json:"value"`
	RawValue string              `json:"raw_value,omitempty"`
}

type CommandArgs struct {
	UserId          string             `json:"user_id"`
	ChannelId       string             `json:"channel_id"`
	TeamId          string             `json:"team_id"`
	RootId          string             `json:"root_id"`
	ParentId        string             `json:"parent_id"`
	TriggerId       string             `json:"trigger_id,omitempty"`
	ConnectionId    string             `json:"connection_id,omitempty"`
	Command         string             `json:"command"`
	SiteURL         string             `json:"-"`
	T               i18n.TranslateFunc `json:"-"`
	UserMentions    UserMentionMap     `json:"-"`
	ChannelMentions ChannelMentionMap  `json:"-"`
	// Options contains structured parsed command arguments (Discord/Slack level)
	Options         []CommandOptionArg `json:"options,omitempty"`
	// Parameters provides key-value dictionary access to parsed option values
	Parameters      map[string]any     `json:"parameters,omitempty"`
}

func (o *CommandArgs) Auditable() map[string]any {
	return map[string]any{
		"user_id":       o.UserId,
		"channel_id":    o.ChannelId,
		"team_id":       o.TeamId,
		"root_id":       o.RootId,
		"parent_id":     o.ParentId,
		"trigger_id":    o.TriggerId,
		"connection_id": o.ConnectionId,
		"command":       o.Command,
		"site_url":      o.SiteURL,
	}
}

// AddUserMention adds or overrides an entry in UserMentions with name username
// and identifier userId
func (o *CommandArgs) AddUserMention(username, userId string) {
	if o.UserMentions == nil {
		o.UserMentions = make(UserMentionMap)
	}

	o.UserMentions[username] = userId
}

// AddChannelMention adds or overrides an entry in ChannelMentions with name
// channelName and identifier channelId
func (o *CommandArgs) AddChannelMention(channelName, channelId string) {
	if o.ChannelMentions == nil {
		o.ChannelMentions = make(ChannelMentionMap)
	}

	o.ChannelMentions[channelName] = channelId
}

// GetOption returns the option by name (case-insensitive), or nil if not found
func (o *CommandArgs) GetOption(name string) *CommandOptionArg {
	for i := range o.Options {
		if strings.EqualFold(o.Options[i].Name, name) {
			return &o.Options[i]
		}
	}
	return nil
}

// GetStringOption returns string value of the option or defaultVal if not found or not string
func (o *CommandArgs) GetStringOption(name, defaultVal string) string {
	if opt := o.GetOption(name); opt != nil {
		if str, ok := opt.Value.(string); ok {
			return str
		}
	}
	return defaultVal
}

// GetBoolOption returns bool value of the option or defaultVal if not found or not bool
func (o *CommandArgs) GetBoolOption(name string, defaultVal bool) bool {
	if opt := o.GetOption(name); opt != nil {
		if b, ok := opt.Value.(bool); ok {
			return b
		}
	}
	return defaultVal
}

