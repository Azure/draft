// Code generated by MockGen. DO NOT EDIT.
// Source: ./az-client.go
//
// Generated by this command:
//
//	mockgen -source=./az-client.go -destination=./mock/az-client.go .
//

// Package mock_providers is a generated GoMock package.
package mock_providers

import (
	context "context"
	reflect "reflect"

	runtime "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armauthorization "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	armsubscription "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	gomock "go.uber.org/mock/gomock"
)

// MockazTenantClient is a mock of azTenantClient interface.
type MockazTenantClient struct {
	ctrl     *gomock.Controller
	recorder *MockazTenantClientMockRecorder
}

// MockazTenantClientMockRecorder is the mock recorder for MockazTenantClient.
type MockazTenantClientMockRecorder struct {
	mock *MockazTenantClient
}

// NewMockazTenantClient creates a new mock instance.
func NewMockazTenantClient(ctrl *gomock.Controller) *MockazTenantClient {
	mock := &MockazTenantClient{ctrl: ctrl}
	mock.recorder = &MockazTenantClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockazTenantClient) EXPECT() *MockazTenantClientMockRecorder {
	return m.recorder
}

// NewListPager mocks base method.
func (m *MockazTenantClient) NewListPager(options *armsubscription.TenantsClientListOptions) *runtime.Pager[armsubscription.TenantsClientListResponse] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewListPager", options)
	ret0, _ := ret[0].(*runtime.Pager[armsubscription.TenantsClientListResponse])
	return ret0
}

// NewListPager indicates an expected call of NewListPager.
func (mr *MockazTenantClientMockRecorder) NewListPager(options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewListPager", reflect.TypeOf((*MockazTenantClient)(nil).NewListPager), options)
}

// MockGraphClient is a mock of GraphClient interface.
type MockGraphClient struct {
	ctrl     *gomock.Controller
	recorder *MockGraphClientMockRecorder
}

// MockGraphClientMockRecorder is the mock recorder for MockGraphClient.
type MockGraphClientMockRecorder struct {
	mock *MockGraphClient
}

// NewMockGraphClient creates a new mock instance.
func NewMockGraphClient(ctrl *gomock.Controller) *MockGraphClient {
	mock := &MockGraphClient{ctrl: ctrl}
	mock.recorder = &MockGraphClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGraphClient) EXPECT() *MockGraphClientMockRecorder {
	return m.recorder
}

// GetApplicationObjectId mocks base method.
func (m *MockGraphClient) GetApplicationObjectId(ctx context.Context, appId string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationObjectId", ctx, appId)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationObjectId indicates an expected call of GetApplicationObjectId.
func (mr *MockGraphClientMockRecorder) GetApplicationObjectId(ctx, appId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationObjectId", reflect.TypeOf((*MockGraphClient)(nil).GetApplicationObjectId), ctx, appId)
}

// MockRoleAssignClient is a mock of RoleAssignClient interface.
type MockRoleAssignClient struct {
	ctrl     *gomock.Controller
	recorder *MockRoleAssignClientMockRecorder
}

// MockRoleAssignClientMockRecorder is the mock recorder for MockRoleAssignClient.
type MockRoleAssignClientMockRecorder struct {
	mock *MockRoleAssignClient
}

// NewMockRoleAssignClient creates a new mock instance.
func NewMockRoleAssignClient(ctrl *gomock.Controller) *MockRoleAssignClient {
	mock := &MockRoleAssignClient{ctrl: ctrl}
	mock.recorder = &MockRoleAssignClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRoleAssignClient) EXPECT() *MockRoleAssignClientMockRecorder {
	return m.recorder
}

// CreateByID mocks base method.
func (m *MockRoleAssignClient) CreateByID(ctx context.Context, roleAssignmentID string, parameters armauthorization.RoleAssignmentCreateParameters, options *armauthorization.RoleAssignmentsClientCreateByIDOptions) (armauthorization.RoleAssignmentsClientCreateByIDResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateByID", ctx, roleAssignmentID, parameters, options)
	ret0, _ := ret[0].(armauthorization.RoleAssignmentsClientCreateByIDResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateByID indicates an expected call of CreateByID.
func (mr *MockRoleAssignClientMockRecorder) CreateByID(ctx, roleAssignmentID, parameters, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateByID", reflect.TypeOf((*MockRoleAssignClient)(nil).CreateByID), ctx, roleAssignmentID, parameters, options)
}
