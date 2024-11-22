package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"

	"github.com/Azure/draft/pkg/prompts"
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
	TenantId          string
	appObjectId       string
	spObjectId        string
	AzClient          AzClientInterface
}

const CONTRIBUTOR_ROLE_ID = "b24988ac-6180-42a0-ab88-20f7382dd24c"

func InitiateAzureOIDCFlow(ctx context.Context, sc *SetUpCmd, s spinner.Spinner, gh GhClient, az AzClientInterface) error {
	log.Debug("Commencing github connection with azure...")

	s.Start()

	if err := sc.ValidateSetUpConfig(gh, az); err != nil {
		return err
	}

	if !az.AzAppExists(sc.AppName) {
		err := az.CreateAzApp(sc.AppName)
		if err != nil {
			return err
		}
	}

	if err := az.CreateServicePrincipal(sc.appId); err != nil {
		return err
	}

	if err := sc.getAppObjectId(); err != nil {
		return err
	}

	if err := az.AssignSpRole(ctx, sc.SubscriptionID, sc.ResourceGroupName, sc.spObjectId, CONTRIBUTOR_ROLE_ID); err != nil {
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

func (az *AzClient) CreateAzApp(appName string) error {
	log.Debug("Commencing Azure app creation...")
	start := time.Now()
	log.Debug(start)

	createApp := func() error {
		out, err := az.CommandRunner.RunCommand("az", "ad", "app", "create", "--only-show-errors", "--display-name", appName)
		if err != nil {
			log.Printf("%s\n", out)
			return err
		}

		if az.AzAppExists(appName) {
			var azApp map[string]interface{}
			if err := json.Unmarshal([]byte(out), &azApp); err != nil {
				return err
			}
			createdAppId := fmt.Sprint(azApp["appId"])

			end := time.Since(start)
			log.Debugf("App with appId '%s' created successfully!", createdAppId)
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

func (az *AzClient) CreateServicePrincipal(appId string) error {
	log.Debug("creating Azure service principal...")
	start := time.Now()
	log.Debug(start)

	createServicePrincipal := func() error {
		out, err := az.CommandRunner.RunCommand("az", "ad", "sp", "create", "--id", appId, "--only-show-errors")
		if err != nil {
			log.Printf("%s\n", out)
			return err
		}

		log.Debug("checking sp was created...")
		if az.ServicePrincipalExists(appId) {
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

// Prompt the user to select a tenant ID if there are multiple tenants, or return the only tenant ID if there is only one
func PromptTenantId(azc AzClientInterface, ctx context.Context) (string, error) {
	log.Debug("getting Azure tenant ID")

	selectedTenant := ""
	tenants, err := azc.ListTenants(ctx)
	if err != nil {
		return selectedTenant, fmt.Errorf("listing tenants: %w", err)
	}

	if len(tenants) == 0 {
		return selectedTenant, errors.New("no tenants found")
	}

	if len(tenants) == 1 {
		if tenants[0].TenantID == nil {
			return selectedTenant, errors.New("nil tenant ID")
		}
		selectedTenant = *tenants[0].TenantID
		log.Debugf("Selected only tenant ID found: %s", selectedTenant)
		return selectedTenant, nil
	}
	if len(tenants) > 1 {
		prompts.Select[armsubscription.TenantIDDescription]("Select the tenant you want to use", tenants, &prompts.SelectOpt[armsubscription.TenantIDDescription]{})
	}

	return selectedTenant, nil
}

func (sc *SetUpCmd) ValidateSetUpConfig(gh GhClient, az AzClientInterface) error {
	log.Debug("Checking that provided information is valid...")

	if err := az.IsSubscriptionIdValid(sc.SubscriptionID); err != nil {
		return err
	}

	if err := az.IsValidResourceGroup(sc.SubscriptionID, sc.ResourceGroupName); err != nil {
		return err
	}

	if sc.AppName == "" {
		return errors.New("invalid app name")
	}

	if err := gh.IsValidGhRepo(sc.Repo); err != nil {
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

func (sc *SetUpCmd) getAppObjectId() error {
	log.Debug("Fetching Azure application object ID")

	getObjectIdCmd := exec.Command("az", "ad", "app", "show", "--only-show-errors", "--id", sc.appId, "--query", "id")
	out, err := getObjectIdCmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", out)
		return err
	}

	var objectId string
	if err := json.Unmarshal(out, &objectId); err != nil {
		return err
	}

	sc.appObjectId = objectId

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
	setTenantIdCmd := exec.Command("gh", "secret", "set", "AZURE_TENANT_ID", "-b", sc.TenantId, "--repo", sc.Repo)
	out, err := setTenantIdCmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", out)
		return err
	}
	return err
}
