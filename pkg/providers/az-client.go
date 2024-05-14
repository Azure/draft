package providers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	graphapp "github.com/microsoftgraph/msgraph-sdk-go/applications"
)

type AzClient struct {
	AzTenantClient azTenantClient
	GraphClient    GraphClient
}

//go:generate mockgen -source=./az-client.go -destination=./mock/az-client.go .
type azTenantClient interface {
	NewListPager(options *armsubscription.TenantsClientListOptions) *runtime.Pager[armsubscription.TenantsClientListResponse]
}

type GraphClient interface {
	Applications() *graphapp.ApplicationsRequestBuilder
}

var _ GraphClient = &msgraph.GraphServiceClient{}

func GetApplicationObjectId(ctx context.Context, appId string, graphClient GraphClient) (string, error) {
	req := graphClient.Applications().ByApplicationId(appId)

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
