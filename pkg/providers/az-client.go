package providers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	log "github.com/sirupsen/logrus"
)

type AzClient struct {
	AzTenantClient   azTenantClient
	GraphClient      GraphClient
	RoleAssignClient RoleAssignClient
}

//go:generate mockgen -source=./az-client.go -destination=./mock/az-client.go .
type azTenantClient interface {
	NewListPager(options *armsubscription.TenantsClientListOptions) *runtime.Pager[armsubscription.TenantsClientListResponse]
}

// GraphServiceClient implements the GraphClient interface.
type GraphServiceClient struct {
	Client *msgraph.GraphServiceClient
}

type GraphClient interface {
	GetApplicationObjectId(ctx context.Context, appId string) (string, error)
}

var _ GraphClient = &GraphServiceClient{}

func (g *GraphServiceClient) GetApplicationObjectId(ctx context.Context, appId string) (string, error) {
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

type RoleAssignmentClient struct {
	Client *armauthorization.RoleAssignmentsClient
}

type RoleAssignClient interface {
	CreateRoleAssignment(ctx context.Context, objectId, roleId, scope, raUid string) error
}

var _ RoleAssignClient = &RoleAssignmentClient{}

func (r *RoleAssignmentClient) CreateRoleAssignment(ctx context.Context, objectId, roleId, scope, raUid string) error {
	log.Debug("Assigning contributor role to service principal...", "objectId", objectId, "role assignment UID", raUid, "scope", scope)

	fullAssignmentId := fmt.Sprintf("/%s/providers/Microsoft.Authorization/roleAssignments/%s", scope, raUid)
	fullDefinitionId := fmt.Sprintf("/providers/Microsoft.Authorization/roleDefinitions/%s", roleId)

	params := armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      &objectId,
			RoleDefinitionID: &fullDefinitionId,
		},
	}

	resp, err := r.Client.CreateByID(ctx, fullAssignmentId, params, nil)
	log.Debug("response from create role assignment", "resp", resp)

	if err != nil {
		return fmt.Errorf("creating role assignment: %w", err)
	}

	return nil
}
