// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

// AutocompleteArgType describes autocomplete argument type
type AutocompleteArgType string

// Argument types
const (
	AutocompleteArgTypeText        AutocompleteArgType = "TextInput"
	AutocompleteArgTypeStaticList  AutocompleteArgType = "StaticList"
	AutocompleteArgTypeDynamicList AutocompleteArgType = "DynamicList"

	// Rich argument types (Discord/Slack level)
	AutocompleteArgTypeInteger AutocompleteArgType = "Integer"
	AutocompleteArgTypeNumber  AutocompleteArgType = "Number"
	AutocompleteArgTypeBoolean AutocompleteArgType = "Boolean"
	AutocompleteArgTypeUser    AutocompleteArgType = "User"
	AutocompleteArgTypeChannel AutocompleteArgType = "Channel"
	AutocompleteArgTypeChoice  AutocompleteArgType = "Choice"
)

// AutocompleteData describes slash command autocomplete information.
type AutocompleteData struct {
	// Trigger of the command
	Trigger string
	// Hint of a command
	Hint string
	// Text displayed to the user to help with the autocomplete description
	HelpText string
	// Role of the user who should be able to see the autocomplete info of this command
	RoleID string
	// Arguments of the command. Arguments can be named or positional.
	// If they are positional order in the list matters, if they are named order does not matter.
	// All arguments should be either named or positional, no mixing allowed.
	Arguments []*AutocompleteArg
	// Subcommands of the command
	SubCommands []*AutocompleteData
}

// AutocompleteArg describes an argument of the command. Arguments can be named or positional.
// If Name is empty string Argument is positional otherwise it is named argument.
// Named arguments are passed as --Name Argument_Value.
type AutocompleteArg struct {
	// Name of the argument
	Name string
	// Text displayed to the user to help with the autocomplete
	HelpText string
	// Type of the argument
	Type AutocompleteArgType
	// Required determines if argument is optional or not.
	Required bool
	// Actual data of the argument (depends on the Type)
	Data any
}

// AutocompleteTextArg describes text user can input as an argument.
type AutocompleteTextArg struct {
	// Hint of the input text
	Hint string
	// Regex pattern to match
	Pattern string
}

// AutocompleteListItem describes an item in the AutocompleteStaticListArg.
type AutocompleteListItem struct {
	Item     string
	Hint     string
	HelpText string
}

// AutocompleteStaticListArg is used to input one of the arguments from the list,
// for example [yes, no], [on, off], and so on.
type AutocompleteStaticListArg struct {
	PossibleArguments []AutocompleteListItem
}

// AutocompleteDynamicListArg is used when user wants to download possible argument list from the URL.
type AutocompleteDynamicListArg struct {
	FetchURL string
}

// AutocompleteChoice describes a structured choice option (label and value, Discord style).
type AutocompleteChoice struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// AutocompleteChoiceArg is used for Choice arguments with a static list of selectable choices.
type AutocompleteChoiceArg struct {
	Choices []AutocompleteChoice `json:"choices"`
}

// AutocompleteSuggestion describes a single suggestion item sent to the front-end
// Example: for user input `/jira cre` -
// Complete might be `/jira create`
// Suggestion might be `create`,
// Hint might be `[issue text]`,
// Description might be `Create a new Issue`
type AutocompleteSuggestion struct {
	// Complete describes completed suggestion
	Complete string
	// Suggestion describes what user might want to input next
	Suggestion string
	// Hint describes a hint about the suggested input
	Hint string
	// Description of the command or a suggestion
	Description string
	// IconData is base64 encoded svg image
	IconData string
}

// NewAutocompleteData returns new Autocomplete data.
func NewAutocompleteData(trigger, hint, helpText string) *AutocompleteData {
	return &AutocompleteData{
		Trigger:     trigger,
		Hint:        hint,
		HelpText:    helpText,
		RoleID:      SystemUserRoleId,
		Arguments:   []*AutocompleteArg{},
		SubCommands: []*AutocompleteData{},
	}
}

// AddCommand adds a subcommand to the autocomplete data.
func (ad *AutocompleteData) AddCommand(command *AutocompleteData) {
	ad.SubCommands = append(ad.SubCommands, command)
}

// AddTextArgument adds positional AutocompleteArgTypeText argument to the command.
func (ad *AutocompleteData) AddTextArgument(helpText, hint, pattern string) {
	ad.AddNamedTextArgument("", helpText, hint, pattern, true)
}

// AddNamedTextArgument adds named AutocompleteArgTypeText argument to the command.
func (ad *AutocompleteData) AddNamedTextArgument(name, helpText, hint, pattern string, required bool) {
	argument := AutocompleteArg{
		Name:     name,
		HelpText: helpText,
		Type:     AutocompleteArgTypeText,
		Required: required,
		Data:     &AutocompleteTextArg{Hint: hint, Pattern: pattern},
	}
	ad.Arguments = append(ad.Arguments, &argument)
}

// AddStaticListArgument adds positional AutocompleteArgTypeStaticList argument to the command.
func (ad *AutocompleteData) AddStaticListArgument(helpText string, required bool, items []AutocompleteListItem) {
	ad.AddNamedStaticListArgument("", helpText, required, items)
}

// AddNamedStaticListArgument adds named AutocompleteArgTypeStaticList argument to the command.
func (ad *AutocompleteData) AddNamedStaticListArgument(name, helpText string, required bool, items []AutocompleteListItem) {
	argument := AutocompleteArg{
		Name:     name,
		HelpText: helpText,
		Type:     AutocompleteArgTypeStaticList,
		Required: required,
		Data:     &AutocompleteStaticListArg{PossibleArguments: items},
	}
	ad.Arguments = append(ad.Arguments, &argument)
}

// AddDynamicListArgument adds positional AutocompleteArgTypeDynamicList argument to the command.
func (ad *AutocompleteData) AddDynamicListArgument(helpText, url string, required bool) {
	ad.AddNamedDynamicListArgument("", helpText, url, required)
}

// AddNamedDynamicListArgument adds named AutocompleteArgTypeDynamicList argument to the command.
func (ad *AutocompleteData) AddNamedDynamicListArgument(name, helpText, url string, required bool) {
	argument := AutocompleteArg{
		Name:     name,
		HelpText: helpText,
		Type:     AutocompleteArgTypeDynamicList,
		Required: required,
		Data:     &AutocompleteDynamicListArg{FetchURL: url},
	}
	ad.Arguments = append(ad.Arguments, &argument)
}

// AddUserArgument adds positional AutocompleteArgTypeUser argument to the command.
func (ad *AutocompleteData) AddUserArgument(helpText, hint string, required bool) {
	ad.AddNamedUserArgument("", helpText, hint, required)
}

// AddNamedUserArgument adds named AutocompleteArgTypeUser argument to the command.
func (ad *AutocompleteData) AddNamedUserArgument(name, helpText, hint string, required bool) {
	ad.Arguments = append(ad.Arguments, &AutocompleteArg{
		Name:     name,
		HelpText: helpText,
		Type:     AutocompleteArgTypeUser,
		Required: required,
		Data:     &AutocompleteTextArg{Hint: hint},
	})
}

// AddChannelArgument adds positional AutocompleteArgTypeChannel argument to the command.
func (ad *AutocompleteData) AddChannelArgument(helpText, hint string, required bool) {
	ad.AddNamedChannelArgument("", helpText, hint, required)
}

// AddNamedChannelArgument adds named AutocompleteArgTypeChannel argument to the command.
func (ad *AutocompleteData) AddNamedChannelArgument(name, helpText, hint string, required bool) {
	ad.Arguments = append(ad.Arguments, &AutocompleteArg{
		Name:     name,
		HelpText: helpText,
		Type:     AutocompleteArgTypeChannel,
		Required: required,
		Data:     &AutocompleteTextArg{Hint: hint},
	})
}

// AddBooleanArgument adds positional AutocompleteArgTypeBoolean argument to the command.
func (ad *AutocompleteData) AddBooleanArgument(helpText string, required bool) {
	ad.AddNamedBooleanArgument("", helpText, required)
}

// AddNamedBooleanArgument adds named AutocompleteArgTypeBoolean argument to the command.
func (ad *AutocompleteData) AddNamedBooleanArgument(name, helpText string, required bool) {
	ad.Arguments = append(ad.Arguments, &AutocompleteArg{
		Name:     name,
		HelpText: helpText,
		Type:     AutocompleteArgTypeBoolean,
		Required: required,
		Data:     nil,
	})
}

// AddIntegerArgument adds positional AutocompleteArgTypeInteger argument to the command.
func (ad *AutocompleteData) AddIntegerArgument(helpText, hint string, required bool) {
	ad.AddNamedIntegerArgument("", helpText, hint, required)
}

// AddNamedIntegerArgument adds named AutocompleteArgTypeInteger argument to the command.
func (ad *AutocompleteData) AddNamedIntegerArgument(name, helpText, hint string, required bool) {
	ad.Arguments = append(ad.Arguments, &AutocompleteArg{
		Name:     name,
		HelpText: helpText,
		Type:     AutocompleteArgTypeInteger,
		Required: required,
		Data:     &AutocompleteTextArg{Hint: hint},
	})
}

// AddChoiceArgument adds positional AutocompleteArgTypeChoice argument to the command.
func (ad *AutocompleteData) AddChoiceArgument(helpText string, required bool, choices []AutocompleteChoice) {
	ad.AddNamedChoiceArgument("", helpText, required, choices)
}

// AddNamedChoiceArgument adds named AutocompleteArgTypeChoice argument to the command.
func (ad *AutocompleteData) AddNamedChoiceArgument(name, helpText string, required bool, choices []AutocompleteChoice) {
	ad.Arguments = append(ad.Arguments, &AutocompleteArg{
		Name:     name,
		HelpText: helpText,
		Type:     AutocompleteArgTypeChoice,
		Required: required,
		Data:     &AutocompleteChoiceArg{Choices: choices},
	})
}

// Equals method checks if command is the same.
func (ad *AutocompleteData) Equals(command *AutocompleteData) bool {
	if !(ad.Trigger == command.Trigger && ad.HelpText == command.HelpText && ad.RoleID == command.RoleID && ad.Hint == command.Hint) {
		return false
	}
	if len(ad.Arguments) != len(command.Arguments) || len(ad.SubCommands) != len(command.SubCommands) {
		return false
	}
	for i := range ad.Arguments {
		if !ad.Arguments[i].Equals(command.Arguments[i]) {
			return false
		}
	}
	for i := range ad.SubCommands {
		if !ad.SubCommands[i].Equals(command.SubCommands[i]) {
			return false
		}
	}
	return true
}

// UpdateRelativeURLsForPluginCommands method updates relative urls for plugin commands
func (ad *AutocompleteData) UpdateRelativeURLsForPluginCommands(baseURL *url.URL) error {
	for _, arg := range ad.Arguments {
		if arg.Type != AutocompleteArgTypeDynamicList {
			continue
		}
		dynamicList, ok := arg.Data.(*AutocompleteDynamicListArg)
		if !ok {
			return errors.New("Not a proper DynamicList type argument")
		}
		dynamicListURL, err := url.Parse(dynamicList.FetchURL)
		if err != nil {
			return errors.Wrapf(err, "FetchURL is not a proper url")
		}
		if !dynamicListURL.IsAbs() {
			absURL := &url.URL{}
			*absURL = *baseURL
			absURL.Path = path.Join(absURL.Path, dynamicList.FetchURL)
			dynamicList.FetchURL = absURL.String()
		}
	}
	for _, command := range ad.SubCommands {
		err := command.UpdateRelativeURLsForPluginCommands(baseURL)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsValid method checks if autocomplete data is valid.
func (ad *AutocompleteData) IsValid() error {
	if ad == nil {
		return errors.New("No nil commands are allowed in AutocompleteData")
	}
	if ad.Trigger == "" {
		return errors.New("An empty command name in the autocomplete data")
	}
	if strings.ToLower(ad.Trigger) != ad.Trigger {
		return errors.New("Command should be lowercase")
	}
	roles := []string{SystemAdminRoleId, SystemUserRoleId, ""}
	if !slices.Contains(roles, ad.RoleID) {
		return errors.New("Wrong role in the autocomplete data")
	}
	if len(ad.Arguments) > 0 && len(ad.SubCommands) > 0 {
		return errors.New("Command can't have arguments and subcommands")
	}
	if len(ad.Arguments) > 0 {
		namedArgumentIndex := -1
		for i, arg := range ad.Arguments {
			if arg.Name != "" { // it's a named argument
				if namedArgumentIndex == -1 { // first named argument
					namedArgumentIndex = i
				}
			} else { // it's a positional argument
				if namedArgumentIndex != -1 {
					return errors.New("Named argument should not be before positional argument")
				}
			}
			if arg.Type == AutocompleteArgTypeDynamicList {
				dynamicList, ok := arg.Data.(*AutocompleteDynamicListArg)
				if !ok {
					return errors.New("Not a proper DynamicList type argument")
				}
				_, err := url.Parse(dynamicList.FetchURL)
				if err != nil {
					return errors.Wrapf(err, "FetchURL is not a proper url")
				}
			} else if arg.Type == AutocompleteArgTypeStaticList {
				staticList, ok := arg.Data.(*AutocompleteStaticListArg)
				if !ok {
					return errors.New("Not a proper StaticList type argument")
				}
				for _, arg := range staticList.PossibleArguments {
					if arg.Item == "" {
						return errors.New("Possible argument name not set in StaticList argument")
					}
				}
			} else if arg.Type == AutocompleteArgTypeText {
				if _, ok := arg.Data.(*AutocompleteTextArg); !ok {
					return errors.New("Not a proper TextInput type argument")
				}
				if arg.Name == "" && !arg.Required {
					return errors.New("Positional argument can not be optional")
				}
			} else if arg.Type == AutocompleteArgTypeChoice {
				choiceArg, ok := arg.Data.(*AutocompleteChoiceArg)
				if !ok || len(choiceArg.Choices) == 0 {
					return errors.New("Choice argument must have non-empty Choices")
				}
			}
		}
	}
	for _, command := range ad.SubCommands {
		err := command.IsValid()
		if err != nil {
			return err
		}
	}
	return nil
}

// Equals method checks if argument is the same.
func (a *AutocompleteArg) Equals(arg *AutocompleteArg) bool {
	if a.Name != arg.Name ||
		a.HelpText != arg.HelpText ||
		a.Type != arg.Type ||
		a.Required != arg.Required ||
		!reflect.DeepEqual(a.Data, arg.Data) {
		return false
	}
	return true
}

// UnmarshalJSON will unmarshal argument
func (a *AutocompleteArg) UnmarshalJSON(b []byte) error {
	var arg map[string]any
	if err := json.Unmarshal(b, &arg); err != nil {
		return errors.Wrapf(err, "Can't unmarshal argument %s", string(b))
	}
	var ok bool
	a.Name, ok = arg["Name"].(string)
	if !ok {
		return errors.Errorf("No field Name in the argument %s", string(b))
	}

	a.HelpText, ok = arg["HelpText"].(string)
	if !ok {
		return errors.Errorf("No field HelpText in the argument %s", string(b))
	}

	t, ok := arg["Type"].(string)
	if !ok {
		return errors.Errorf("No field Type in the argument %s", string(b))
	}
	a.Type = AutocompleteArgType(t)

	a.Required, ok = arg["Required"].(bool)
	if !ok {
		return errors.Errorf("No field Required in the argument %s", string(b))
	}

	data, ok := arg["Data"]
	if !ok && a.Type != AutocompleteArgTypeBoolean {
		return errors.Errorf("No field Data in the argument %s", string(b))
	}

	if a.Type == AutocompleteArgTypeText {
		m, ok := data.(map[string]any)
		if !ok {
			return errors.Errorf("Wrong Data type in the TextInput argument %s", string(b))
		}
		pattern, ok := m["Pattern"].(string)
		if !ok {
			return errors.Errorf("No field Pattern in the TextInput argument %s", string(b))
		}
		hint, ok := m["Hint"].(string)
		if !ok {
			return errors.Errorf("No field Hint in the TextInput argument %s", string(b))
		}
		a.Data = &AutocompleteTextArg{Hint: hint, Pattern: pattern}
	} else if a.Type == AutocompleteArgTypeStaticList {
		m, ok := data.(map[string]any)
		if !ok {
			return errors.Errorf("Wrong Data type in the StaticList argument %s", string(b))
		}
		list, ok := m["PossibleArguments"].([]any)
		if !ok {
			return errors.Errorf("No field PossibleArguments in the StaticList argument %s", string(b))
		}

		possibleArguments := []AutocompleteListItem{}
		for i := range list {
			args, ok := list[i].(map[string]any)
			if !ok {
				return errors.Errorf("Wrong AutocompleteStaticListItem type in the StaticList argument %s", string(b))
			}
			item, ok := args["Item"].(string)
			if !ok {
				return errors.Errorf("No field Item in the StaticList's possible arguments %s", string(b))
			}

			hint, ok := args["Hint"].(string)
			if !ok {
				return errors.Errorf("No field Hint in the StaticList's possible arguments %s", string(b))
			}
			helpText, ok := args["HelpText"].(string)
			if !ok {
				return errors.Errorf("No field Hint in the StaticList's possible arguments %s", string(b))
			}

			possibleArguments = append(possibleArguments, AutocompleteListItem{
				Item:     item,
				Hint:     hint,
				HelpText: helpText,
			})
		}
		a.Data = &AutocompleteStaticListArg{PossibleArguments: possibleArguments}
	} else if a.Type == AutocompleteArgTypeDynamicList {
		m, ok := data.(map[string]any)
		if !ok {
			return errors.Errorf("Wrong type in the DynamicList argument %s", string(b))
		}
		url, ok := m["FetchURL"].(string)
		if !ok {
			return errors.Errorf("No field FetchURL in the DynamicList's argument %s", string(b))
		}
		a.Data = &AutocompleteDynamicListArg{FetchURL: url}
	} else if a.Type == AutocompleteArgTypeChoice {
		m, ok := data.(map[string]any)
		if !ok {
			return errors.Errorf("Wrong type in the Choice argument %s", string(b))
		}
		rawChoices, ok := m["Choices"].([]any)
		if !ok {
			return errors.Errorf("No field Choices in the Choice argument %s", string(b))
		}
		choices := make([]AutocompleteChoice, 0, len(rawChoices))
		for _, c := range rawChoices {
			cm, ok := c.(map[string]any)
			if !ok {
				continue
			}
			name, _ := cm["name"].(string)
			if name == "" {
				name, _ = cm["Name"].(string)
			}
			val := cm["value"]
			if val == nil {
				val = cm["Value"]
			}
			choices = append(choices, AutocompleteChoice{Name: name, Value: val})
		}
		a.Data = &AutocompleteChoiceArg{Choices: choices}
	} else if a.Type == AutocompleteArgTypeUser || a.Type == AutocompleteArgTypeChannel || a.Type == AutocompleteArgTypeInteger || a.Type == AutocompleteArgTypeNumber {
		if m, ok := data.(map[string]any); ok {
			hint, _ := m["Hint"].(string)
			pattern, _ := m["Pattern"].(string)
			a.Data = &AutocompleteTextArg{Hint: hint, Pattern: pattern}
		}
	} else if a.Type == AutocompleteArgTypeBoolean {
		a.Data = nil
	}
	return nil
}

// TokenizeCommandLine splits a raw command line into tokens respecting single and double quotes.
func TokenizeCommandLine(cmd string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range cmd {
		if inQuote {
			if r == quoteChar {
				inQuote = false
			} else {
				current.WriteRune(r)
			}
		} else {
			if r == '"' || r == '\'' {
				inQuote = true
				quoteChar = r
			} else if unicode.IsSpace(r) {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func convertOptionValue(argType AutocompleteArgType, raw string, data any) (any, error) {
	switch argType {
	case AutocompleteArgTypeBoolean:
		if strings.EqualFold(raw, "true") || raw == "1" || strings.EqualFold(raw, "yes") || strings.EqualFold(raw, "y") {
			return true, nil
		} else if strings.EqualFold(raw, "false") || raw == "0" || strings.EqualFold(raw, "no") || strings.EqualFold(raw, "n") {
			return false, nil
		}
		return false, errors.Errorf("expected boolean value (true/false), got '%s'", raw)
	case AutocompleteArgTypeInteger:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, errors.Errorf("expected integer value, got '%s'", raw)
		}
		return v, nil
	case AutocompleteArgTypeNumber:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, errors.Errorf("expected number value, got '%s'", raw)
		}
		return v, nil
	case AutocompleteArgTypeUser:
		return strings.TrimPrefix(raw, "@"), nil
	case AutocompleteArgTypeChannel:
		return strings.TrimPrefix(raw, "~"), nil
	case AutocompleteArgTypeChoice:
		if choiceArg, ok := data.(*AutocompleteChoiceArg); ok && choiceArg != nil {
			for _, choice := range choiceArg.Choices {
				if strings.EqualFold(choice.Name, raw) || strings.EqualFold(fmt.Sprintf("%v", choice.Value), raw) {
					return choice.Value, nil
				}
			}
		}
		return raw, nil
	default:
		return raw, nil
	}
}

// ParseArguments parses a raw arguments string according to the AutocompleteData definition.
func (ad *AutocompleteData) ParseArguments(rawArgs string) ([]CommandOptionArg, map[string]any, error) {
	if ad == nil || len(ad.Arguments) == 0 {
		return nil, nil, nil
	}

	tokens := TokenizeCommandLine(strings.TrimSpace(rawArgs))
	options := make([]CommandOptionArg, 0, len(ad.Arguments))
	params := make(map[string]any)

	isNamed := false
	for _, arg := range ad.Arguments {
		if arg.Name != "" {
			isNamed = true
			break
		}
	}

	if isNamed {
		argMap := make(map[string]*AutocompleteArg)
		for _, arg := range ad.Arguments {
			argMap[strings.ToLower(arg.Name)] = arg
		}

		for i := 0; i < len(tokens); i++ {
			token := tokens[i]
			if strings.HasPrefix(token, "--") {
				flagName := strings.TrimPrefix(token, "--")
				var flagValue string
				hasExplicitValue := false

				if eqIdx := strings.Index(flagName, "="); eqIdx != -1 {
					flagValue = flagName[eqIdx+1:]
					flagName = flagName[:eqIdx]
					hasExplicitValue = true
				}

				def, ok := argMap[strings.ToLower(flagName)]
				if !ok {
					continue
				}

				if def.Type == AutocompleteArgTypeBoolean && !hasExplicitValue {
					if i+1 < len(tokens) && (strings.EqualFold(tokens[i+1], "true") || strings.EqualFold(tokens[i+1], "false")) {
						flagValue = tokens[i+1]
						i++
					} else {
						flagValue = "true"
					}
				} else if !hasExplicitValue {
					if i+1 < len(tokens) {
						flagValue = tokens[i+1]
						i++
					}
				}

				val, err := convertOptionValue(def.Type, flagValue, def.Data)
				if err != nil {
					return nil, nil, errors.Wrapf(err, "invalid value for parameter '--%s'", def.Name)
				}

				opt := CommandOptionArg{
					Name:     def.Name,
					Type:     def.Type,
					Value:    val,
					RawValue: flagValue,
				}
				options = append(options, opt)
				params[def.Name] = val
			}
		}

		for _, def := range ad.Arguments {
			if def.Required {
				if _, present := params[def.Name]; !present {
					return nil, nil, errors.Errorf("missing required parameter: '--%s'", def.Name)
				}
			}
		}
	} else {
		tokenIdx := 0
		for _, def := range ad.Arguments {
			if tokenIdx < len(tokens) {
				rawVal := tokens[tokenIdx]
				tokenIdx++

				val, err := convertOptionValue(def.Type, rawVal, def.Data)
				if err != nil {
					return nil, nil, errors.Wrapf(err, "invalid argument at position %d", tokenIdx)
				}

				opt := CommandOptionArg{
					Name:     def.Name,
					Type:     def.Type,
					Value:    val,
					RawValue: rawVal,
				}
				options = append(options, opt)
				if def.Name != "" {
					params[def.Name] = val
				}
			} else if def.Required {
				return nil, nil, errors.Errorf("missing required argument '%s'", def.HelpText)
			}
		}
	}

	return options, params, nil
}
