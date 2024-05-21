package cred

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

var (
	cred      *azidentity.DefaultAzureCredential
	graphCred *azidentity.DeviceCodeCredential
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

func GetGraphCred() (*azidentity.DeviceCodeCredential, error) {
	if graphCred != nil {
		return graphCred, nil
	}

	gCreds, _ := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		TenantID: "", // todo - set this value
		ClientID: "", // todo - set this value
		UserPrompt: func(ctx context.Context, message azidentity.DeviceCodeMessage) error {
			fmt.Printf(message.Message)
			return nil
		},
	})

	return gCreds, nil
}
