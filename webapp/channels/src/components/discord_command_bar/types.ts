// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

export interface CommandChoice {
    name: string;
    value: string;
}

export interface CommandSlot {
    name: string;
    label: string;
    type: 'string' | 'integer' | 'number' | 'boolean' | 'user' | 'channel' | 'choice';
    required: boolean;
    choices?: CommandChoice[];
    hint?: string;
    value: string;
}

export interface ActiveCommandState {
    trigger: string;
    description?: string;
    slots: CommandSlot[];
}
