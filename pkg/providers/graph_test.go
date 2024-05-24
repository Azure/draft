package providers

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
)

func TestLocal(t *testing.T) {

	cred, _ := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		TenantID: "72f988bf-86f1-41af-91ab-2d7cd011db47",
		// ClientID: "60171198-7cdf-40df-9e03-ebf2ab7a8917",
		ClientID: "f1e800fd-fbb1-40a0-a0f2-81eef128e6fd", // draft-test-app
		UserPrompt: func(ctx context.Context, message azidentity.DeviceCodeMessage) error {
			fmt.Println(message.Message)
			return nil
		},
	})

	graphClient, _ := graph.NewGraphServiceClientWithCredentials(
		cred, []string{"Application.Read.All"})

	applications, err := graphClient.Applications().Get(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(applications)

}
