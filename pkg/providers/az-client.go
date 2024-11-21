package providers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

//go:generate mockgen -source=./az-client.go -destination=./mock/az-client.go .
type AzClientInterface interface {
	ListResourceGroups(ctx context.Context, subscriptionID string) ([]armresources.ResourceGroup, error)
	ListTenants(ctx context.Context) ([]armsubscription.TenantIDDescription, error)
	assignSpRole(ctx context.Context, subscriptionId, resourceGroup, servicePrincipalObjectID, roleId string) error
}

// assert AzClient implements AzClientInterface
var _ AzClientInterface = &AzClient{}

// AzClient is a struct that contains the Azure client and its dependencies
// It is used to interact with Azure resources
// Create a new AzClient with NewAzClient
type AzClient struct {
	Credential          *azidentity.DefaultAzureCredential
	TenantClient        *armsubscription.TenantsClient
	RoleAssignClient    *armauthorization.RoleAssignmentsClient
	ResourceGroupClient *armresources.ResourceGroupsClient
}

func NewAzClient(cred *azidentity.DefaultAzureCredential) (*AzClient, error) {
	azClient := &AzClient{
		Credential: cred,
	}
	return azClient, nil
}

func (az *AzClient) ListResourceGroups(ctx context.Context, subscriptionID string) ([]armresources.ResourceGroup, error) {
	log.Debug("listing Azure resource groups for subscription ", subscriptionID)
	if az.ResourceGroupClient == nil {
		c, err := armresources.NewResourceGroupsClient(subscriptionID, az.Credential, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource group client: %w", err)
		}
		az.ResourceGroupClient = c
	}

	var rgs []armresources.ResourceGroup
	pager := az.ResourceGroupClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing resource groups page: %w", err)
		}

		for _, rg := range page.Value {
			if rg == nil {
				return nil, errors.New("nil rg")
			}

			rgs = append(rgs, *rg)
		}
	}

	return rgs, nil
}

func (az *AzClient) ListTenants(ctx context.Context) ([]armsubscription.TenantIDDescription, error) {
	log.Debug("Starting to list Azure Tenants")

	// Initialize the tenant slice to store the results.
	tenants := make([]armsubscription.TenantIDDescription, 0)

	if az.TenantClient == nil {
		c, err := armsubscription.NewTenantsClient(az.Credential, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create tenant client: %w", err)
		}
		az.TenantClient = c
	}
	pager := az.TenantClient.NewListPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing tenants page: %w", err)
		}

		for _, t := range page.Value {
			if t == nil {
				return nil, errors.New("nil tenant") // this should never happen but it's good to check just in case
			}
			tenants = append(tenants, *t)
		}
	}

	log.Debugf("Successfully listed %d Azure tenants", len(tenants))
	return tenants, nil
}

func (az *AzClient) assignSpRole(ctx context.Context, subscriptionId, resourceGroup, servicePrincipalObjectID, roleId string) error {
	log.Debug("Assigning contributor role to service principal...")
	if az.RoleAssignClient == nil {
		c, err := armauthorization.NewRoleAssignmentsClient(subscriptionId, az.Credential, nil)
		if err != nil {
			return fmt.Errorf("failed to create role assignment client: %w", err)
		}
		az.RoleAssignClient = c
	}

	scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionId, resourceGroup)
	objectID := servicePrincipalObjectID
	raUid := uuid.New().String()

	fullAssignmentId := fmt.Sprintf("/%s/providers/Microsoft.Authorization/roleAssignments/%s", scope, raUid)
	fullDefinitionId := fmt.Sprintf("/providers/Microsoft.Authorization/roleDefinitions/%s", roleId)

	principalType := armauthorization.PrincipalTypeServicePrincipal
	parameters := armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      &objectID,
			RoleDefinitionID: &fullDefinitionId,
			PrincipalType:    &principalType,
		},
	}

	_, err := az.RoleAssignClient.CreateByID(ctx, fullAssignmentId, parameters, nil)
	if err != nil {
		return fmt.Errorf("creating role assignment: %w", err)
	}

	log.Debug("Role assigned successfully!")
	return nil
}
