package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/providers"
	"github.com/Azure/draft/pkg/spinner"
)

func TestSetUpConfig(t *testing.T) {
	mockSetUpCmd := &providers.SetUpCmd{}
	mockSetUpCmd.AppName = "testingSetUpCommand"
	mockSetUpCmd.Provider = "Google"
	mockSetUpCmd.Repo = "test/repo"
	mockSetUpCmd.ResourceGroupName = "testResourceGroup"
	mockSetUpCmd.SubscriptionID = "123456789"
	s := spinner.CreateSpinner("--> Setting up Github OIDC...")

	fillSetUpConfig(mockSetUpCmd)

	err := runProviderSetUp(mockSetUpCmd, s)

	assert.True(t, err == nil)
}