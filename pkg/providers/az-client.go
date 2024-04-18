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
	AzTenantClient azTenantClient
}

//go:generate mockgen -source=./az_client.go -destination=./mock/az_client.go .
type azTenantClient interface {
	NewListPager(options *armsubscription.TenantsClientListOptions) *runtime.Pager[armsubscription.TenantsClientListResponse]
}

type GraphClient interface {
	GetApplicationObjectId(ctx context.Context, appId string) (string, error)
}

// GraphServiceClientImpl implements the GraphClient interface.
type GraphServiceClientImpl struct {
	Client *msgraph.GraphServiceClient
}

func (g *GraphServiceClientImpl) GetApplicationObjectId(ctx context.Context, appId string) (string, error) {
	req := g.Client.Applications().ByApplicationId(appId)

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
