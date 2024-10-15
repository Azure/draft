package providers

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/Azure/draft/pkg/cred"
)

type AzClient struct {
	AzTenantClient   azTenantClient
	RoleAssignClient *armauthorization.RoleAssignmentsClient
}

//go:generate mockgen -source=./az-client.go -destination=./mock/az-client.go .
type azTenantClient interface {
	NewListPager(options *armsubscription.TenantsClientListOptions) *runtime.Pager[armsubscription.TenantsClientListResponse]
}

func createRoleAssignmentClient(subscriptionId string) (*armauthorization.RoleAssignmentsClient, error) {
	credential, err := cred.GetCred()
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	client, err := armauthorization.NewRoleAssignmentsClient(subscriptionId, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create role assignment client: %w", err)
	}
	return client, nil
}
