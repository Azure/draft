package providers

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	graphapp "github.com/microsoftgraph/msgraph-sdk-go/applications"

	mock_providers "github.com/Azure/draft/pkg/providers/mock"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

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

var testAppID = "mockAppID"
var testID = "mockID"
var errToSend error = nil

type mockApplicationable struct {
	models.Applicationable
}

func (m *mockApplicationable) GetAppId() *string {
	return &testAppID
}

func (m *mockApplicationable) GetId() *string {
	return &testID
}

type mockSerialWriter struct {
	serialization.SerializationWriter
}

func (m *mockSerialWriter) GetSerializedContent() ([]byte, error) {
	content := []byte("a few bytes")
	return content, nil
}

func (m *mockSerialWriter) Close() error {
	return nil
}

func (m *mockSerialWriter) WriteObjectValue(string, serialization.Parsable, ...serialization.Parsable) error {
	return nil
}

type mockSerialWriterFactory struct {
	serialization.SerializationWriterFactory
}

func (m *mockSerialWriterFactory) GetSerializationWriter(string) (serialization.SerializationWriter, error) {
	return &mockSerialWriter{}, nil
}

type mockRequestAdapter struct {
	abstractions.RequestAdapter
}

func (m *mockRequestAdapter) Send(
	context.Context,
	*abstractions.RequestInformation,
	serialization.ParsableFactory,
	abstractions.ErrorMappings) (serialization.Parsable, error) {
	return &mockApplicationable{}, errToSend
}

func (m *mockRequestAdapter) GetSerializationWriterFactory() serialization.SerializationWriterFactory {
	return &mockSerialWriterFactory{}
}

func TestGetAppObjectId(t *testing.T) {
	tests := []struct {
		name      string
		testAppID string
		errToSend error
		expectErr bool
	}{
		{
			name:      "Success",
			testAppID: "mockAppID",
			errToSend: nil,
			expectErr: false,
		},
		{
			name:      "EmptyAppID",
			testAppID: "",
			errToSend: nil,
			expectErr: true,
		},
		{
			name:      "ErrorFromGraphClient",
			testAppID: "testAppID",
			errToSend: errors.New("getting application object Id: mock error"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGraphClient := mock_providers.NewMockGraphClient(ctrl)

			mockAppRequestBuilder := &graphapp.ApplicationsRequestBuilder{
				BaseRequestBuilder: abstractions.BaseRequestBuilder{
					PathParameters: map[string]string{"key": "value"},
					RequestAdapter: &mockRequestAdapter{},
					UrlTemplate:    "dummyUrlTemplate",
				},
			}

			testAppID = tt.testAppID
			errToSend = tt.errToSend
			mockGraphClient.EXPECT().Applications().Return(mockAppRequestBuilder).AnyTimes()

			sc := &SetUpCmd{
				appId: tt.testAppID,
				AzClient: AzClient{
					GraphClient: mockGraphClient,
				},
			}

			err := sc.getAppObjectId(context.Background())
			if tt.expectErr && err == nil {
				t.Error("Expected an error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

var principalId = "mockPrincipalID"
var roleDefId = "mockRoleDefinitionID"
var Id = "mockID"
var name = "mockName"
var Idtype = "mocktype"

func TestAssignSpRole(t *testing.T) {
	tests := []struct {
		name          string
		expectedError error
		mockResponse  armauthorization.RoleAssignmentsClientCreateByIDResponse
	}{
		{
			name:          "Success",
			expectedError: nil,
			mockResponse: armauthorization.RoleAssignmentsClientCreateByIDResponse{
				RoleAssignment: armauthorization.RoleAssignment{
					Properties: &armauthorization.RoleAssignmentPropertiesWithScope{
						PrincipalID:      &principalId,
						RoleDefinitionID: &roleDefId,
					},
					ID:   &Id,
					Name: &name,
					Type: &Idtype,
				},
			},
		},
		{
			name:          "Error",
			expectedError: errors.New("error"),
			mockResponse:  armauthorization.RoleAssignmentsClientCreateByIDResponse{},
		},
		{
			name:          "ErrorDuringRoleAssignment",
			expectedError: errors.New("error during role assignment"),
			mockResponse:  armauthorization.RoleAssignmentsClientCreateByIDResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleAssignClient := mock_providers.NewMockRoleAssignClient(ctrl)

			mockRoleAssignClient.EXPECT().CreateByID(gomock.Any(), "contributor", gomock.Any(), gomock.Any()).Return(tt.mockResponse, tt.expectedError)

			sc := &SetUpCmd{
				AzClient: AzClient{
					RoleAssignClient: mockRoleAssignClient,
				},
				spObjectId: "testObjectId",
			}

			err := sc.assignSpRole(context.Background())
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestCreateAzApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGraphClient := mock_providers.NewMockGraphClient(ctrl)

	mockAppRequestBuilder := &graphapp.ApplicationsRequestBuilder{
		BaseRequestBuilder: abstractions.BaseRequestBuilder{
			PathParameters: map[string]string{"key": "value"},
			RequestAdapter: &mockRequestAdapter{},
			UrlTemplate:    "dummyUrlTemplate",
		},
	}

	mockGraphClient.EXPECT().Applications().Return(mockAppRequestBuilder).AnyTimes()

	sc := &SetUpCmd{
		AzClient: AzClient{
			GraphClient: mockGraphClient,
		},
		AppName: "AppName",
	}

	err := sc.createAzApp(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestCreateAzApp_ErrorCreatingApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGraphClient := mock_providers.NewMockGraphClient(ctrl)

	errToSend = errors.New("getting application object Id: mock error")

	mockAppRequestBuilder := &graphapp.ApplicationsRequestBuilder{
		BaseRequestBuilder: abstractions.BaseRequestBuilder{
			PathParameters: map[string]string{"key": "value"},
			RequestAdapter: &mockRequestAdapter{},
			UrlTemplate:    "dummyUrlTemplate",
		},
	}

	expectedErr := errors.New("creating Azure app: getting application object Id: mock error")

	mockGraphClient.EXPECT().Applications().Return(mockAppRequestBuilder).AnyTimes()

	sc := &SetUpCmd{
		AzClient: AzClient{
			GraphClient: mockGraphClient,
		},
		AppName: "",
	}

	err := sc.createAzApp(context.Background())

	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Expected error %v, got: %v", expectedErr, err)
	}
}
