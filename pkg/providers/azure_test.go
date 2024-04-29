package providers

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	mock_providers "github.com/Azure/draft/pkg/providers/mock"
	"go.uber.org/mock/gomock"
	"strings"
	"testing"
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

func TestGetAppObjectId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGraphClient := mock_providers.NewMockGraphClient(ctrl)

	appID := "testAppID"
	expectedAppID := "mockAppID"
	mockGraphClient.EXPECT().GetApplicationObjectId(gomock.Any(), appID).Return(expectedAppID, nil)

	sc := &SetUpCmd{
		appId: appID,
		AzClient: AzClient{
			GraphClient: mockGraphClient,
		},
	}

	err := sc.getAppObjectId(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if sc.appObjectId != expectedAppID {
		t.Errorf("Expected application ID %s, got: %s", expectedAppID, sc.appObjectId)
	}
}

func TestGetAppObjectId_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGraphClient := mock_providers.NewMockGraphClient(ctrl)

	appID := "testAppID"
	expectedError := errors.New("mock error")
	mockGraphClient.EXPECT().GetApplicationObjectId(gomock.Any(), appID).Return("", expectedError)

	sc := &SetUpCmd{
		appId: appID,
		AzClient: AzClient{
			GraphClient: mockGraphClient,
		},
	}

	err := sc.getAppObjectId(context.Background())

	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

// Test case - when the GraphClient returns an error
func TestGetAppObjectId_ErrorFromGraphClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGraphClient := mock_providers.NewMockGraphClient(ctrl)

	appID := "testAppID"
	expectedError := errors.New("mock error")
	mockGraphClient.EXPECT().GetApplicationObjectId(gomock.Any(), appID).Return("", expectedError)

	sc := &SetUpCmd{
		appId: appID,
		AzClient: AzClient{
			GraphClient: mockGraphClient,
		},
	}

	err := sc.getAppObjectId(context.Background())
	if err == nil {
		t.Error("Expected an error, got nil")
	}
	expectedErrorMsg := "getting application object Id: mock error"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// Test case - when the GraphClient returns an empty application ID:
func TestGetAppObjectId_EmptyAppIdFromGraphClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGraphClient := mock_providers.NewMockGraphClient(ctrl)

	appID := "testAppID"
	expectedError := errors.New("application object ID is empty")
	mockGraphClient.EXPECT().GetApplicationObjectId(gomock.Any(), appID).Return("", expectedError)

	sc := &SetUpCmd{
		appId: appID,
		AzClient: AzClient{
			GraphClient: mockGraphClient,
		},
	}

	err := sc.getAppObjectId(context.Background())

	if err == nil {
		t.Error("Expected an error, got nil")
	} else if !strings.Contains(err.Error(), expectedError.Error()) {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}
}

func TestAssignSpRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleAssignClient := mock_providers.NewMockRoleAssignClient(ctrl)

	expectedObjectId := "testObjectId"
	expectedRoleId := "contributor"
	expectedScope := "/subscriptions/testSubscriptionID/resourceGroups/testResourceGroupName"
	mockRoleAssignClient.EXPECT().CreateRoleAssignment(gomock.Any(), expectedObjectId, expectedRoleId, expectedScope, gomock.Any()).Return(nil)

	sc := &SetUpCmd{
		AzClient: AzClient{
			RoleAssignClient: mockRoleAssignClient,
		},
		SubscriptionID:    "testSubscriptionID",
		ResourceGroupName: "testResourceGroupName",
		spObjectId:        expectedObjectId,
	}

	err := sc.assignSpRole(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestAssignSpRole_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleAssignClient := mock_providers.NewMockRoleAssignClient(ctrl)

	expectedObjectId := "testObjectId"
	expectedRoleId := "contributor"
	expectedScope := "/subscriptions/testSubscriptionID/resourceGroups/testResourceGroupName"
	expectedError := errors.New("error")

	mockRoleAssignClient.EXPECT().CreateRoleAssignment(gomock.Any(), expectedObjectId, expectedRoleId, expectedScope, gomock.Any()).Return(expectedError)

	sc := &SetUpCmd{
		AzClient: AzClient{
			RoleAssignClient: mockRoleAssignClient,
		},
		SubscriptionID:    "testSubscriptionID",
		ResourceGroupName: "testResourceGroupName",
		spObjectId:        expectedObjectId,
	}

	err := sc.assignSpRole(context.Background())
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestAssignSpRole_ErrorDuringRoleAssignment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleAssignClient := mock_providers.NewMockRoleAssignClient(ctrl)

	expectedObjectId := "testObjectId"
	expectedRoleId := "contributor"
	expectedScope := "/subscriptions/testSubscriptionID/resourceGroups/testResourceGroupName"
	expectedError := errors.New("error during role assignment")

	mockRoleAssignClient.EXPECT().CreateRoleAssignment(gomock.Any(), expectedObjectId, expectedRoleId, expectedScope, gomock.Any()).Return(expectedError)

	sc := &SetUpCmd{
		AzClient: AzClient{
			RoleAssignClient: mockRoleAssignClient,
		},
		SubscriptionID:    "testSubscriptionID",
		ResourceGroupName: "testResourceGroupName",
		spObjectId:        expectedObjectId,
	}

	err := sc.assignSpRole(context.Background())
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
}
