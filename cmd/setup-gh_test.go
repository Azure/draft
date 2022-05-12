package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/providers"
)

func TestSetUpConfig(t *testing.T) {
	mockSetUpCmd := &providers.SetUpCmd{}
	mockSetUpCmd.AppName = "testingSetUpCommand"
	mockSetUpCmd.Provider = "Google"
	mockSetUpCmd.Repo = "test/repo"
	mockSetUpCmd.ResourceGroupName = "testResourceGroup"
	mockSetUpCmd.SubscriptionID = "123456789"

	fillSetUpConfig(mockSetUpCmd)

	err := runProviderSetUp(mockSetUpCmd)

	assert.True(t, err == nil)
}