package providers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
)

type AzClient struct {
	AzTenantClient     azTenantClient
	GraphServiceClient GraphServiceClient
}

//go:generate mockgen -source=./az_client.go -destination=./mock/az_client.go .
type azTenantClient interface {
	NewListPager(options *armsubscription.TenantsClientListOptions) *runtime.Pager[armsubscription.TenantsClientListResponse]
}

// GraphServiceClient implements the GraphClient interface.
type GraphServiceClient struct {
	Client *msgraph.GraphServiceClient
}

type GraphClient interface {
	GetApplicationObjectId(ctx context.Context, appId string, graphServiceClient GraphServiceClient) (string, error)
}

func GetApplicationObjectId(ctx context.Context, appId string, graphServiceClient GraphServiceClient) (string, error) {
	req := graphServiceClient.Client.Applications().ByApplicationId(appId)

	app, err := req.Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("getting application details: %w", err)
	}
	appObjectId := app.GetAppId()
	if appObjectId == nil || *appObjectId == "" {
		return "", errors.New("application object ID is empty")
	}
	return *appObjectId, nil
}