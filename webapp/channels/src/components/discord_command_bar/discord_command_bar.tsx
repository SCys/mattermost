// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useRef, useEffect, useCallback} from 'react';
import type {KeyboardEvent, ChangeEvent} from 'react';

import type {CommandSlot, CommandChoice} from './types';
import './discord_command_bar.scss';

export interface DiscordCommandBarProps {
    trigger: string;
    description?: string;
    initialSlots: CommandSlot[];
    onCancel: () => void;
    onSubmit: (fullCommandText: string) => void;
    onValueChange?: (fullCommandText: string) => void;
}

export const extractSlotsFromHint = (hint: string): CommandSlot[] => {
    if (!hint || !hint.trim()) {
        return [];
    }

    const slots: CommandSlot[] = [];
    const tokens = hint.match(/\[[^\]]+\]|<[^>]+>|--[a-zA-Z0-9_-]+/g) || [];

    for (const rawToken of tokens) {
        const isRequired = rawToken.startsWith('<') && rawToken.endsWith('>');
        let clean = rawToken.trim();
        if ((clean.startsWith('[') && clean.endsWith(']')) || (clean.startsWith('<') && clean.endsWith('>'))) {
            clean = clean.slice(1, -1).trim();
        }
        clean = clean.replace(/\*$/, '').trim();

        if (clean.includes(':')) {
            const [paramName, typeOrChoices] = clean.split(':').map((s) => s.trim());
            if (typeOrChoices.includes('|')) {
                const choiceList: CommandChoice[] = typeOrChoices.split('|').map((c) => ({
                    name: c.trim(),
                    value: c.trim(),
                }));
                slots.push({
                    name: paramName,
                    label: paramName,
                    type: 'choice',
                    required: isRequired,
                    choices: choiceList,
                    value: '',
                });
                continue;
            }

            if (typeOrChoices.toLowerCase() === 'boolean' || typeOrChoices.toLowerCase() === 'bool') {
                slots.push({
                    name: paramName,
                    label: paramName,
                    type: 'boolean',
                    required: isRequired,
                    choices: [
                        {name: 'true', value: 'true'},
                        {name: 'false', value: 'false'},
                    ],
                    value: '',
                });
                continue;
            }

            slots.push({
                name: paramName,
                label: paramName,
                type: 'string',
                required: isRequired,
                value: '',
            });
        } else if (clean.startsWith('--')) {
            const flagName = clean.replace(/^--/, '');
            slots.push({
                name: flagName,
                label: flagName,
                type: 'boolean',
                required: isRequired,
                choices: [
                    {name: 'true', value: 'true'},
                    {name: 'false', value: 'false'},
                ],
                value: '',
            });
        } else if (clean.startsWith('@')) {
            const name = clean.slice(1).replace(/^[<\[]+|[>\]]+$/g, '').trim();
            slots.push({
                name: name || 'user',
                label: name || 'user',
                type: 'user',
                required: isRequired,
                value: '',
            });
        } else if (clean.startsWith('~')) {
            const name = clean.slice(1).replace(/^[<\[]+|[>\]]+$/g, '').trim();
            slots.push({
                name: name || 'channel',
                label: name || 'channel',
                type: 'channel',
                required: isRequired,
                value: '',
            });
        } else {
            slots.push({
                name: clean,
                label: clean,
                type: 'string',
                required: isRequired,
                value: '',
            });
        }
    }

    return slots;
};

export const DiscordCommandBar: React.FC<DiscordCommandBarProps> = ({
    trigger,
    description,
    initialSlots,
    onCancel,
    onSubmit,
    onValueChange,
}) => {
    const [slots, setSlots] = useState<CommandSlot[]>(initialSlots);
    const [activeSlotIndex, setActiveSlotIndex] = useState<number>(0);
    const [dropdownIndex, setDropdownIndex] = useState<number>(0);
    const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

    const buildCommandString = useCallback((currentSlots: CommandSlot[]): string => {
        let cmd = `/${trigger}`;
        for (const slot of currentSlots) {
            if (slot.value && slot.value.trim()) {
                const isPositional = ['args', 'text', 'message', 'words', '文字', '内容'].includes(slot.name.toLowerCase());
                if (isPositional) {
                    cmd += ` ${slot.value.trim()}`;
                } else if (slot.type === 'boolean' && slot.value === 'true') {
                    cmd += ` --${slot.name}`;
                } else {
                    cmd += ` --${slot.name} ${slot.value.trim()}`;
                }
            }
        }
        return cmd;
    }, [trigger]);

    useEffect(() => {
        if (inputRefs.current[activeSlotIndex]) {
            inputRefs.current[activeSlotIndex]?.focus();
        }
    }, [activeSlotIndex]);

    const handleSlotChange = (index: number, val: string) => {
        const next = [...slots];
        next[index] = {...next[index], value: val};
        setSlots(next);
        setDropdownIndex(0);

        const cmdStr = buildCommandString(next);
        onValueChange?.(cmdStr);
    };

    const handleSelectChoice = (index: number, choiceValue: string) => {
        const next = [...slots];
        next[index] = {...next[index], value: choiceValue};
        setSlots(next);

        const cmdStr = buildCommandString(next);
        onValueChange?.(cmdStr);

        if (index < slots.length - 1) {
            setActiveSlotIndex(index + 1);
        }
    };

    const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>, index: number) => {
        const currentSlot = slots[index];
        const choices = currentSlot.choices || (currentSlot.type === 'boolean' ? [
            {name: 'true', value: 'true'},
            {name: 'false', value: 'false'},
        ] : undefined);

        if (choices && choices.length > 0) {
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                setDropdownIndex((prev) => (prev + 1) % choices.length);
                return;
            }
            if (e.key === 'ArrowUp') {
                e.preventDefault();
                setDropdownIndex((prev) => (prev - 1 + choices.length) % choices.length);
                return;
            }
            if (e.key === 'Enter' && choices[dropdownIndex]) {
                e.preventDefault();
                handleSelectChoice(index, choices[dropdownIndex].value);
                return;
            }
        }

        if (e.key === 'Tab') {
            e.preventDefault();
            if (e.shiftKey) {
                if (index > 0) {
                    setActiveSlotIndex(index - 1);
                }
            } else if (index < slots.length - 1) {
                setActiveSlotIndex(index + 1);
            }
            return;
        }

        if (e.key === 'Backspace' && !currentSlot.value) {
            if (index === 0) {
                e.preventDefault();
                onCancel();
            } else {
                e.preventDefault();
                setActiveSlotIndex(index - 1);
            }
            return;
        }

        if (e.key === 'Escape') {
            e.preventDefault();
            onCancel();
            return;
        }

        if (e.key === 'Enter') {
            e.preventDefault();
            const cmdStr = buildCommandString(slots);
            onSubmit(cmdStr);
        }
    };

    return (
        <div className='discord-command-bar'>
            <div className='discord-command-bar__command-pill'>
                <span className='discord-command-bar__icon'>/</span>
                <span className='discord-command-bar__trigger'>{trigger}</span>
                <span
                    className='discord-command-bar__close-btn'
                    onClick={onCancel}
                    title='取消命令 (Cancel command)'
                >
                    ✕
                </span>
            </div>

            <div className='discord-command-bar__slots-container'>
                {slots.map((slot, index) => {
                    const isActive = activeSlotIndex === index;
                    const isFilled = Boolean(slot.value);
                    const choices = slot.choices || (slot.type === 'boolean' ? [
                        {name: 'true', value: 'true'},
                        {name: 'false', value: 'false'},
                    ] : undefined);

                    return (
                        <div
                            key={slot.name}
                            className={`discord-command-bar__slot ${isActive ? 'discord-command-bar__slot--active' : ''} ${isFilled ? 'discord-command-bar__slot--filled' : ''} ${slot.required ? 'discord-command-bar__slot--required' : ''}`}
                            onClick={() => setActiveSlotIndex(index)}
                        >
                            <span className='discord-command-bar__slot-label'>
                                {slot.label}
                                {slot.required && <span className='discord-command-bar__req-star'>*</span>}
                            </span>
                            <input
                                ref={(el) => {
                                    inputRefs.current[index] = el;
                                }}
                                type='text'
                                className='discord-command-bar__slot-input'
                                value={slot.value}
                                placeholder={slot.hint || (choices ? '选择...' : '输入...')}
                                style={{
                                    width: `${Math.max(slot.value.length * 8 + 10, 36)}px`,
                                }}
                                onChange={(e: ChangeEvent<HTMLInputElement>) => handleSlotChange(index, e.target.value)}
                                onKeyDown={(e: KeyboardEvent<HTMLInputElement>) => handleKeyDown(e, index)}
                            />

                            {isActive && choices && choices.length > 0 && (
                                <div className='discord-command-bar__dropdown'>
                                    {choices.map((choice, cIdx) => (
                                        <div
                                            key={choice.value}
                                            className={`discord-command-bar__dropdown-item ${dropdownIndex === cIdx ? 'discord-command-bar__dropdown-item--active' : ''}`}
                                            onMouseDown={(e) => {
                                                e.preventDefault();
                                                handleSelectChoice(index, choice.value);
                                            }}
                                        >
                                            <span>{choice.name}</span>
                                            <span className='discord-command-bar__dropdown-value'>{choice.value}</span>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>

            <div className='discord-command-bar__helper-hint'>
                {description && <span title={description} style={{maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap'}}>{description}</span>}
                <span><kbd>Tab</kbd> 下一个</span>
                <span><kbd>Enter</kbd> 发送</span>
                <span><kbd>Esc</kbd> 取消</span>
            </div>
        </div>
    );
};

export default DiscordCommandBar;
