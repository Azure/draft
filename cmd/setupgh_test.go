package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/providers"
	"github.com/Azure/draft/pkg/spinner"
)

func TestSetUpConfig(t *testing.T) {
	ctx := context.Background()
	mockSetUpCmd := &providers.SetUpCmd{}
	mockSetUpCmd.AppName = "testingSetUpCommand"
	mockSetUpCmd.Provider = "Google"
	mockSetUpCmd.Repo = "test/repo"
	mockSetUpCmd.ResourceGroupName = "testResourceGroup"
	mockSetUpCmd.SubscriptionID = "123456789"
	mockSetUpCmd.TenantId = "123456789"
	s := spinner.CreateSpinner("--> Setting up Github OIDC...")

	gh := &providers.GhCliClient{}
	az := &providers.AzClient{}
	fillSetUpConfig(mockSetUpCmd, gh, az)

	err := runProviderSetUp(ctx, mockSetUpCmd, s, gh, az)

	assert.True(t, err == nil)
}

func TestToValidAppName(t *testing.T) {
	testCases := []struct {
		testCaseName string
		nameInput    string
		expected     string
		shouldError  bool
	}{
		{
			testCaseName: "valid name",
			nameInput:    "valid-name",
			expected:     "valid-name",
		},
		{
			testCaseName: "name with spaces",
			nameInput:    "name with spaces",
			expected:     "name-with-spaces",
		},
		{
			testCaseName: "name with special characters",
			nameInput:    "name!@#$%^&*()",
			expected:     "name",
		},
		{
			testCaseName: "cannot start with a period",
			nameInput:    ".name",
			expected:     "name",
		},
		{
			testCaseName: "cannot start or end with hyphen",
			nameInput:    "----name--",
			expected:     "name",
		},
		{
			testCaseName: "name that can't be made valid",
			expected:     ".**(__)",
			shouldError:  true,
		},
		{
			testCaseName: "hypens allowed in the middle",
			nameInput:    "name-name-name-name",
			expected:     "name-name-name-name",
		}, {
			testCaseName: "remove dots in the middle",
			nameInput:    "name.name-name-name",
			expected:     "namename-name-name",
		},
		{
			testCaseName: "no capital letters",
			nameInput:    "NaMe",
			expected:     "name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testCaseName, func(t *testing.T) {
			result, err := toValidAppName(tc.nameInput)
			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.expected, result)
				if err != nil {
					t.Errorf("expected no error, got %s", err)
				}
			}
		})
	}

}

func TestValidateAppName(t *testing.T) {
	cases := []struct {
		name          string
		input         string
		expectedError bool
	}{
		{
			name:  "valid name",
			input: "valid-name",
		},
		{
			name:          "name with spaces",
			input:         "name with spaces",
			expectedError: true,
		},
		{
			name:          "name with special characters",
			input:         "name!@#$%^&*()",
			expectedError: true,
		},
		{
			name:          "cannot start with a period",
			input:         ".name",
			expectedError: true,
		},
		{
			name:          "cannot start or end with hyphen",
			input:         "----name--",
			expectedError: true,
		},
		{
			name:          "cannot end with a period",
			input:         "name-1-a.",
			expectedError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAppName(tc.input)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
