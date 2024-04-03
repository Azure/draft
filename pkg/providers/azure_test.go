package providers

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/tracing"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	mock_providers "github.com/Azure/draft/pkg/providers/mock"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestGetTenantId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testId := "00000000-0000-0000-0000-000000000000"
	testTenantId := "/tenants/00000000-0000-0000-0000-000000000000"
	testNextLink := "https://pkg.go.dev/github.com"
	testTenantDesc := armsubscription.TenantIDDescription{ID: &testId, TenantID: &testTenantId}
	testTenantDescArray := []*armsubscription.TenantIDDescription{&testTenantDesc}
	testTenantListResult := armsubscription.TenantListResult{NextLink: &testNextLink, Value: testTenantDescArray}
	responses := []armsubscription.TenantsClientListResponse{{testTenantListResult}}
	testReadResponses := 0
	testPagerHandler := runtime.PagingHandler[armsubscription.TenantsClientListResponse]{
		More: func(t armsubscription.TenantsClientListResponse) bool { return testReadResponses < len(responses) },
		Fetcher: func(ctx context.Context, response *armsubscription.TenantsClientListResponse) (armsubscription.TenantsClientListResponse, error) {
			resp := responses[testReadResponses]
			testReadResponses++
			return resp, nil
		},
		Tracer: tracing.Tracer{},
	}
	mockClient := mock_providers.NewMockazTenantClient(ctrl)

	var mockPager = runtime.NewPager[armsubscription.TenantsClientListResponse](testPagerHandler)

	mockClient.EXPECT().NewListPager(gomock.Nil()).Return(mockPager).Times(1)

	sc := &SetUpCmd{
		AzTenantClient: mockClient,
	}

	err := sc.getTenantId(context.Background())
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
