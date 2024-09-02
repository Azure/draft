package providers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	mock_providers "github.com/bfoley13/draft/pkg/providers/mock"
	"go.uber.org/mock/gomock"
)

func setupMockClientAndPager(ctrl *gomock.Controller, responses []armsubscription.TenantsClientListResponse) *mock_providers.MockazTenantClient {
	mockClient := mock_providers.NewMockazTenantClient(ctrl)

	// Define a minimal paging handler function that returns the provided responses
	mockPagerHandler := runtime.PagingHandler[armsubscription.TenantsClientListResponse]{
		More: func(t armsubscription.TenantsClientListResponse) bool { return false },
		Fetcher: func(ctx context.Context, response *armsubscription.TenantsClientListResponse) (armsubscription.TenantsClientListResponse, error) {
			if len(responses) == 0 {
				return armsubscription.TenantsClientListResponse{}, nil
			}
			resp := responses[0]
			responses = responses[1:]
			return resp, nil
		},
		Tracer: tracing.Tracer{},
	}

	// Create a mock pager with the paging handler
	mockPager := runtime.NewPager[armsubscription.TenantsClientListResponse](mockPagerHandler)

	mockClient.EXPECT().NewListPager(gomock.Nil()).Return(mockPager).Times(1)

	return mockClient
}

func TestGetTenantId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Define test data
	testId := "00000000-0000-0000-0000-000000000000"
	testTenantId := "/tenants/00000000-0000-0000-0000-000000000000"
	testNextLink := "https://pkg.go.dev/github.com"
	testTenantDesc := armsubscription.TenantIDDescription{ID: &testId, TenantID: &testTenantId}
	testTenantDescArray := []*armsubscription.TenantIDDescription{&testTenantDesc}
	testTenantListResult := armsubscription.TenantListResult{NextLink: &testNextLink, Value: testTenantDescArray}
	responses := []armsubscription.TenantsClientListResponse{{testTenantListResult}}

	mockClient := setupMockClientAndPager(ctrl, responses)

	sc := &SetUpCmd{
		AzClient: AzClient{
			AzTenantClient: mockClient,
		},
	}

	err := sc.getTenantId(context.Background())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Test case for the getTenantId function when listing tenants encounters an error
func TestGetTenantId_ListTenantsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock client and pager to return an error when listing tenants
	mockClient := mock_providers.NewMockazTenantClient(ctrl)
	mockPager := runtime.NewPager[armsubscription.TenantsClientListResponse](runtime.PagingHandler[armsubscription.TenantsClientListResponse]{
		More: func(t armsubscription.TenantsClientListResponse) bool { return false },
		Fetcher: func(ctx context.Context, response *armsubscription.TenantsClientListResponse) (armsubscription.TenantsClientListResponse, error) {
			return armsubscription.TenantsClientListResponse{}, errors.New("error listing tenants")
		},
		Tracer: tracing.Tracer{},
	})
	mockClient.EXPECT().NewListPager(gomock.Nil()).Return(mockPager).Times(1)

	sc := &SetUpCmd{
		AzClient: AzClient{
			AzTenantClient: mockClient,
		},
	}

	err := sc.getTenantId(context.Background())

	if err == nil || !strings.Contains(err.Error(), "error listing tenants") {
		t.Errorf("Expected error listing tenants, got: %v", err)
	}
}

// Test case for the getTenantId function when tenant list is empty
func TestGetTenantId_EmptyTenantList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock client and pager with no responses
	mockClient := setupMockClientAndPager(ctrl, []armsubscription.TenantsClientListResponse{})

	sc := &SetUpCmd{
		AzClient: AzClient{
			AzTenantClient: mockClient,
		},
	}

	err := sc.getTenantId(context.Background())

	if err == nil || !strings.Contains(err.Error(), "no tenants found") {
		t.Errorf("Expected error no tenants found, got: %v", err)
	}
}

// Test case for the getTenantId function when a nil tenant is encountered in the list
func TestGetTenantId_NilTenantInList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Define test data with a nil tenant in the list
	testTenantDescArray := []*armsubscription.TenantIDDescription{nil}

	mockClient := setupMockClientAndPager(ctrl, []armsubscription.TenantsClientListResponse{{TenantListResult: armsubscription.TenantListResult{Value: testTenantDescArray}}})

	sc := &SetUpCmd{
		AzClient: AzClient{
			AzTenantClient: mockClient,
		},
	}

	err := sc.getTenantId(context.Background())

	if err == nil || !strings.Contains(err.Error(), "nil tenant") {
		t.Errorf("Expected error nil tenant, got: %v", err)
	}
}
