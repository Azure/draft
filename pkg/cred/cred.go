package cred

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

var (
	cred *azidentity.DefaultAzureCredential
)

func GetCred() (*azidentity.DefaultAzureCredential, error) {
	if cred != nil {
		return cred, nil
	}

	var err error
	cred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("authenticating to Azure: %w", err)
	}

	return cred, nil
}
