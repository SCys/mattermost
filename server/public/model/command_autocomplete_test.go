// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutocompleteData(t *testing.T) {
	ad := NewAutocompleteData("jira", "", "Available commands:")
	assert.NoError(t, ad.IsValid())
	ad.RoleID = "some_id"
	assert.Error(t, ad.IsValid())
	ad.RoleID = SystemAdminRoleId
	assert.NoError(t, ad.IsValid())
	ad.AddDynamicListArgument("help", "/some/url", true)
	assert.NoError(t, ad.IsValid())
	ad.AddNamedTextArgument("name", "help", "[text]", "", true)
	assert.NoError(t, ad.IsValid())

	ad = getAutocompleteData()
	assert.NoError(t, ad.IsValid())
	command := NewAutocompleteData("", "", "")
	ad.AddCommand(command)
	assert.Error(t, ad.IsValid())

	ad = getAutocompleteData()
	command = NewAutocompleteData("disconnect", "", "disconnect")
	command.AddTextArgument("help", "[text]", "")
	command.AddNamedTextArgument("some", "help", "[text]", "", true)
	ad.AddCommand(command)
	assert.NoError(t, ad.IsValid())

	ad = getAutocompleteData()
	command = NewAutocompleteData("disconnect", "", "disconnect")
	command.AddDynamicListArgument("help", "valid_url", true)
	ad.AddCommand(command)
	assert.NoError(t, ad.IsValid())

	ad = getAutocompleteData()
	command = NewAutocompleteData("disconnect", "", "disconnect")
	command.AddDynamicListArgument("help", "/valid/url", true)
	items := []AutocompleteListItem{
		{
			Hint:     "help",
			Item:     "",
			HelpText: "text",
		},
	}
	command.AddStaticListArgument("help", true, items)
	ad.AddCommand(command)
	assert.Error(t, ad.IsValid())

	ad = getAutocompleteData()
	ad.AddCommand(nil)
	assert.Error(t, ad.IsValid())

	ad = getAutocompleteData()
	command = NewAutocompleteData("Disconnect", "", "")
	ad.AddCommand(command)
	assert.Error(t, ad.IsValid())
}

func getAutocompleteData() *AutocompleteData {
	ad := NewAutocompleteData("jira", "", "Available commands:")
	ad.RoleID = SystemUserRoleId
	command := NewAutocompleteData("connect", "", "Connect to mattermost")
	command.RoleID = SystemAdminRoleId
	items := []AutocompleteListItem{
		{
			Hint:     "arg1",
			Item:     "help1",
			HelpText: "text1",
		}, {
			Hint:     "arg2",
			Item:     "help2",
			HelpText: "text2",
		},
	}
	command.AddStaticListArgument("help", true, items)
	command.AddNamedTextArgument("some", "help", "[text]", "", true)
	command.AddNamedDynamicListArgument("other", "help", "/other/url", true)
	ad.AddCommand(command)
	return ad
}

func TestUpdateRelativeURLsForPluginCommands(t *testing.T) {
	ad := getAutocompleteData()
	baseURL, _ := url.Parse("http://localhost:8065/plugins/com.mattermost.demo-plugin")
	err := ad.UpdateRelativeURLsForPluginCommands(baseURL)
	assert.NoError(t, err)
	arg, ok := ad.SubCommands[0].Arguments[2].Data.(*AutocompleteDynamicListArg)
	assert.True(t, ok)
	assert.Equal(t, "http://localhost:8065/plugins/com.mattermost.demo-plugin/other/url", arg.FetchURL)
}

func TestRichSlashCommandArguments(t *testing.T) {
	deployCmd := NewAutocompleteData("deploy", "", "Deploy service to an environment")
	deployCmd.AddNamedChoiceArgument("env", "Deployment environment", true, []AutocompleteChoice{
		{Name: "Production", Value: "prod"},
		{Name: "Staging", Value: "staging"},
	})
	deployCmd.AddNamedBooleanArgument("dry-run", "Simulate without executing", false)
	deployCmd.AddNamedIntegerArgument("replicas", "Number of replicas", "[count]", false)
	deployCmd.AddNamedUserArgument("approver", "Approving user", "@username", false)

	assert.NoError(t, deployCmd.IsValid())

	// Test ParseArguments: valid input
	rawInput := `--env prod --dry-run true --replicas 3 --approver @alice`
	opts, params, err := deployCmd.ParseArguments(rawInput)
	assert.NoError(t, err)
	assert.Len(t, opts, 4)
	assert.Equal(t, "prod", params["env"])
	assert.Equal(t, true, params["dry-run"])
	assert.Equal(t, int64(3), params["replicas"])
	assert.Equal(t, "alice", params["approver"])

	// Test ParseArguments: boolean flag shorthand
	rawInput2 := `--env staging --dry-run`
	opts2, params2, err := deployCmd.ParseArguments(rawInput2)
	assert.NoError(t, err)
	assert.Equal(t, "staging", params2["env"])
	assert.Equal(t, true, params2["dry-run"])
	_ = opts2

	// Test ParseArguments: missing required parameter
	rawInputMissing := `--dry-run true`
	_, _, err = deployCmd.ParseArguments(rawInputMissing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required parameter: '--env'")
}
