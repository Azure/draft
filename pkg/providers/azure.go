package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"os/exec"
	"time"

	"github.com/Azure/draft/pkg/spinner"

	bo "github.com/cenkalti/backoff/v4"
	log "github.com/sirupsen/logrus"
)

type SetUpCmd struct {
	AppName           string
	SubscriptionID    string
	ResourceGroupName string
	Provider          string
	Repo              string
	appId             string
	tenantId          string
	appObjectId       string
	spObjectId        string
	AzClient          AzClient
}

func InitiateAzureOIDCFlow(ctx context.Context, sc *SetUpCmd, s spinner.Spinner) error {
	log.Debug("Commencing github connection with azure...")

	if !HasGhCli() || !IsLoggedInToGh() {
		s.Stop()
		if err := LogInToGh(); err != nil {
			return err
		}
		s.Start()
	}

	if err := sc.ValidateSetUpConfig(); err != nil {
		return err
	}

	if AzAppExists(sc.AppName) {
		return errors.New("app already exists")
	} else if err := sc.createAzApp(); err != nil {
		return err
	}

	if err := sc.CreateServicePrincipal(); err != nil {
		return err
	}

	if err := sc.getTenantId(ctx); err != nil {
		return err
	}

	if err := sc.getAppObjectId(ctx); err != nil {
		return err
	}

	if err := sc.assignSpRole(ctx); err != nil {
		return err
	}

	if !sc.hasFederatedCredentials() {
		if err := sc.createFederatedCredentials(); err != nil {
			return err
		}
	}

	if err := sc.setAzClientId(); err != nil {
		return err
	}
	if err := sc.setAzSubscriptionId(); err != nil {
		return err
	}
	if err := sc.setAzTenantId(); err != nil {
		return err
	}

	log.Debug("Github connection with azure completed successfully!")
	return nil
}

func (sc *SetUpCmd) createAzApp() error {
	log.Debug("Commencing Azure app creation...")
	start := time.Now()
	log.Debug(start)

	createApp := func() error {
		createAppCmd := exec.Command("az", "ad", "app", "create", "--only-show-errors", "--display-name", sc.AppName)

		out, err := createAppCmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)
			return err
		}

		if AzAppExists(sc.AppName) {
			var azApp map[string]interface{}
			if err := json.Unmarshal(out, &azApp); err != nil {
				return err
			}
			appId := fmt.Sprint(azApp["appId"])

			sc.appId = appId

			end := time.Since(start)
			log.Debug("App created successfully!")
			log.Debug(end)
			return nil
		}

		return errors.New("app creation time has exceeded max elapsed time for exponential backoff")
	}

	backoff := bo.NewExponentialBackOff()
	backoff.MaxElapsedTime = 5 * time.Second

	err := bo.Retry(createApp, backoff)
	if err != nil {
		log.Debug(err)
		return err
	}

	return nil
}

func (sc *SetUpCmd) CreateServicePrincipal() error {
	log.Debug("Creating Azure service principal...")
	start := time.Now()
	log.Debug(start)

	createServicePrincipal := func() error {
		createSpCmd := exec.Command("az", "ad", "sp", "create", "--id", sc.appId, "--only-show-errors")
		out, err := createSpCmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)
			return err
		}

		log.Debug("Checking sp was created...")
		if sc.ServicePrincipalExists() {
			log.Debug("Service principal created successfully!")
			end := time.Since(start)
			log.Debug(end)
			return nil
		}

		return errors.New("service principal not found")
	}

	backoff := bo.NewExponentialBackOff()
	backoff.MaxElapsedTime = 5 * time.Second

	err := bo.Retry(createServicePrincipal, backoff)
	if err != nil {
		log.Debug(err)
		return err
	}

	return nil
}

func (sc *SetUpCmd) assignSpRole(ctx context.Context) error {
	log.Debug("Assigning contributor role to service principal...")

	objectID := sc.spObjectId
	roleID := "contributor"

	parameters := armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      &objectID,
			RoleDefinitionID: &roleID,
		},
	}

	_, err := sc.AzClient.RoleAssignClient.CreateByID(ctx, roleID, parameters, nil)
	if err != nil {
		return fmt.Errorf("creating role assignment: %w", err)
	}

	log.Debug("Role assigned successfully!")
	return nil
}

func (sc *SetUpCmd) getTenantId(ctx context.Context) error {
	log.Debug("getting Azure tenant ID")

	tenants, err := sc.listTenants(ctx)
	if err != nil {
		return fmt.Errorf("listing tenants: %w", err)
	}

	if len(tenants) == 0 {
		return errors.New("no tenants found")
	}
	if len(tenants) > 1 {
		return errors.New("multiple tenants found")
	}
	sc.tenantId = *tenants[0].TenantID

	return nil
}

func (sc *SetUpCmd) listTenants(ctx context.Context) ([]armsubscription.TenantIDDescription, error) {
	log.Debug("listing Azure subscriptions")

	var tenants []armsubscription.TenantIDDescription

	pager := sc.AzClient.AzTenantClient.NewListPager(nil)

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

	log.Debug("finished listing Azure tenants")
	return tenants, nil
}

func (sc *SetUpCmd) ValidateSetUpConfig() error {
	log.Debug("Checking that provided information is valid...")

	if err := IsSubscriptionIdValid(sc.SubscriptionID); err != nil {
		return err
	}

	if err := isValidResourceGroup(sc.SubscriptionID, sc.ResourceGroupName); err != nil {
		return err
	}

	if sc.AppName == "" {
		return errors.New("invalid app name")
	}

	if err := isValidGhRepo(sc.Repo); err != nil {
		return err
	}

	return nil
}

func (sc *SetUpCmd) hasFederatedCredentials() bool {
	log.Debug("Checking for existing federated credentials...")
	uri := fmt.Sprintf("https://graph.microsoft.com/beta/applications/%s/federatedIdentityCredentials", sc.appObjectId)
	getFicCmd := exec.Command("az", "rest", "--method", "GET", "--uri", uri, "--query", "value")
	out, err := getFicCmd.CombinedOutput()
	if err != nil {
		log.Errorf("error getting fic: %s", err)
		return false
	}

	var fics []interface{}
	if err = json.Unmarshal(out, &fics); err != nil {
		log.Errorf("error marshaling fics: %s", err)
		return false
	}

	if len(fics) > 0 {
		log.Debug("Credentials found")
		// TODO: ask user if they want to use current credentials?
		// TODO: check if fics with the name we want exist already
		return true
	}

	log.Debug("No existing credentials found")
	return false
}

func (sc *SetUpCmd) createFederatedCredentials() error {
	log.Debug("Creating federated credentials...")
	fics := &[]string{
		`{"name":"prfic","subject":"repo:%s:pull_request","issuer":"https://token.actions.githubusercontent.com","description":"pr","audiences":["api://AzureADTokenExchange"]}`,
		`{"name":"mainfic","subject":"repo:%s:ref:refs/heads/main","issuer":"https://token.actions.githubusercontent.com","description":"main","audiences":["api://AzureADTokenExchange"]}`,
		`{"name":"masterfic","subject":"repo:%s:ref:refs/heads/master","issuer":"https://token.actions.githubusercontent.com","description":"master","audiences":["api://AzureADTokenExchange"]}`,
	}

	uri := "https://graph.microsoft.com/beta/applications/%s/federatedIdentityCredentials"

	for _, fic := range *fics {
		createFicCmd := exec.Command("az", "rest", "--method", "POST", "--uri", fmt.Sprintf(uri, sc.appObjectId), "--body", fmt.Sprintf(fic, sc.Repo))
		out, err := createFicCmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)
			return err
		}

	}

	log.Debug("Waiting 10 seconds to allow credentials time to populate")
	time.Sleep(10 * time.Second)
	count := 0

	// check to make sure credentials were created
	// count to prevent infinite loop
	for count < 10 {
		if sc.hasFederatedCredentials() {
			break
		}

		log.Debug("Credentials not yet created, retrying...")
		count += 1
	}

	return nil

}

func (sc *SetUpCmd) getAppObjectId(ctx context.Context) error {
	log.Debug("Fetching Azure application object ID")

	appID, err := sc.AzClient.GraphClient.GetApplicationObjectId(ctx, sc.appId)
	if err != nil {
		return fmt.Errorf("getting application object Id: %w", err)
	}

	sc.appObjectId = appID

	return nil
}

func (sc *SetUpCmd) setAzClientId() error {
	log.Debug("Setting AZURE_CLIENT_ID in github...")
	setClientIdCmd := exec.Command("gh", "secret", "set", "AZURE_CLIENT_ID", "-b", sc.appId, "--repo", sc.Repo)
	out, err := setClientIdCmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", out)
		return err
	}
	return nil
}

func (sc *SetUpCmd) setAzSubscriptionId() error {
	log.Debug("Setting AZURE_SUBSCRIPTION_ID in github...")
	setSubscriptionIdCmd := exec.Command("gh", "secret", "set", "AZURE_SUBSCRIPTION_ID", "-b", sc.SubscriptionID, "--repo", sc.Repo)
	out, err := setSubscriptionIdCmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", out)
		return err
	}
	return nil
}

func (sc *SetUpCmd) setAzTenantId() error {
	log.Debug("Setting AZURE_TENANT_ID in github...")
	setTenantIdCmd := exec.Command("gh", "secret", "set", "AZURE_TENANT_ID", "-b", sc.tenantId, "--repo", sc.Repo)
	out, err := setTenantIdCmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", out)
		return err
	}
	return err
}

func (sc *SetUpCmd) CleanUpAzureResources(appName string) error {
	log.Debug("Starting cleanup of Azure resources...")

	// Fetch the app ID using the app name
	if sc.appId == "" {
		getAppIdCmd := exec.Command("az", "ad", "app", "list", "--display-name", appName, "--query", "[0].appId", "--only-show-errors")
		out, err := getAppIdCmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)
			return err
		}

		var appId string
		if err := json.Unmarshal(out, &appId); err != nil {
			return err
		}

		sc.appId = appId
	}

	// Delete federated credentials
	if sc.hasFederatedCredentials() {
		uri := fmt.Sprintf("https://graph.microsoft.com/beta/applications/%s/federatedIdentityCredentials", sc.appObjectId)
		deleteFicCmd := exec.Command("az", "rest", "--method", "DELETE", "--uri", uri)
		out, err := deleteFicCmd.CombinedOutput()
		if err != nil {
			log.Printf("%s\n", out)
			return err
		}
		log.Debug("Deleted federated credentials successfully.")
	}

	// Delete the service principal
	deleteSpCmd := exec.Command("az", "ad", "sp", "delete", "--id", sc.appId, "--only-show-errors")
	out, err := deleteSpCmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", out)
		return err
	}
	log.Debug("Deleted service principal successfully.")

	// Delete the Azure AD application
	deleteAppCmd := exec.Command("az", "ad", "app", "delete", "--id", sc.appId, "--only-show-errors")
	out, err = deleteAppCmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", out)
		return err
	}
	log.Debug("Deleted Azure AD application successfully.")

	return nil
}
