package providers

import (
	"encoding/json"
	"errors"
	"fmt"
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
}

func InitiateAzureOIDCFlow(sc *SetUpCmd, s spinner.Spinner) error {
	log.Debug("Commencing github connection with azure...")

	if !HasGhCli() || !IsLoggedInToGh() {
		s.Stop()
		if err := LogInToGh(); err != nil {
			log.Fatal(err)
		}
		s.Start()
	}

	if err := sc.ValidateSetUpConfig(); err != nil {
		return err
	}

	if AzAppExists(sc.AppName) {
		log.Fatal("App already exists")
	} else if err := sc.createAzApp(); err != nil {
		return err
	}

	if err := sc.CreateServicePrincipal(); err != nil {
		return err
	}

	if err := sc.getTenantId(); err != nil {
		return err
	}

	if err := sc.getAppObjectId(); err != nil {
		return err
	}

	if err := sc.assignSpRole(); err != nil {
		return err
	}

	if !sc.hasFederatedCredentials() {
		if err := sc.createFederatedCredentials(); err != nil {
			return err
		}
	}

	sc.setAzClientId()
	sc.setAzSubscriptionId()
	sc.setAzTenantId()

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
			log.Fatal(out)
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

		return errors.New("app not found")
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
			log.Fatal(out)
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

func (sc *SetUpCmd) assignSpRole() error {
	log.Debug("Assigning contributor role to service principal...")
	scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", sc.SubscriptionID, sc.ResourceGroupName)
	assignSpRoleCmd := exec.Command("az", "role", "assignment", "create", "--role", "contributor", "--subscription", sc.SubscriptionID, "--assignee-object-id", sc.spObjectId, "--assignee-principal-type", "ServicePrincipal", "--scope", scope, "--only-show-errors")
	out, err := assignSpRoleCmd.CombinedOutput()
	if err != nil {
		log.Fatalf(string(out))
		return err
	}

	log.Debug("Role assigned successfully!")
	return nil
}

func (sc *SetUpCmd) getTenantId() error {
	log.Debug("Fetching Azure account tenant ID")
	getTenantIdCmd := exec.Command("az", "account", "show", "--query", "tenantId", "--only-show-errors")
	out, err := getTenantIdCmd.CombinedOutput()
	if err != nil {
		log.Fatalf(string(out))
		return err
	}

	var tenantId string
	if err := json.Unmarshal(out, &tenantId); err != nil {
		return err
	}
	tenantId = fmt.Sprint(tenantId)

	sc.tenantId = tenantId

	return nil
}

func (sc *SetUpCmd) ValidateSetUpConfig() error {
	log.Debug("Checking that provided information is valid...")

	if err := IsSubscriptionIdValid(sc.SubscriptionID); err != nil {
		return err
	}

	if err := isValidResourceGroup(sc.ResourceGroupName); err != nil {
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
		out, ficErr := createFicCmd.CombinedOutput()
		if ficErr != nil {
			log.Fatalf(string(out))
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
		log.Fatalf(string(out))
		return err
	}

	var objectId string
	if err := json.Unmarshal(out, &objectId); err != nil {
		return err
	}

	sc.appObjectId = objectId

	return nil
}

func (sc *SetUpCmd) setAzClientId() {
	log.Debug("Setting AZURE_CLIENT_ID in github...")
	setClientIdCmd := exec.Command("gh", "secret", "set", "AZURE_CLIENT_ID", "-b", sc.appId, "--repo", sc.Repo)
	out, err := setClientIdCmd.CombinedOutput()
	if err != nil {
		log.Fatal(string(out))

	}

}

func (sc *SetUpCmd) setAzSubscriptionId() {
	log.Debug("Setting AZURE_SUBSCRIPTION_ID in github...")
	setSubscriptionIdCmd := exec.Command("gh", "secret", "set", "AZURE_SUBSCRIPTION_ID", "-b", sc.SubscriptionID, "--repo", sc.Repo)
	out, err := setSubscriptionIdCmd.CombinedOutput()
	if err != nil {
		log.Fatal(string(out))

	}

}

func (sc *SetUpCmd) setAzTenantId() {
	log.Debug("Setting AZURE_TENANT_ID in github...")
	setTenantIdCmd := exec.Command("gh", "secret", "set", "AZURE_TENANT_ID", "-b", sc.tenantId, "--repo", sc.Repo)
	out, err := setTenantIdCmd.CombinedOutput()
	if err != nil {
		log.Fatal(string(out))

	}

}
