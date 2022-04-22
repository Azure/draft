package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"time"

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

func InitiateAzureOIDCFlow(sc *SetUpCmd) error {
	log.Debug("Commencing github connection with azure...")

	if !HasGhCli() || !IsLoggedInToGh() {
		if err := LogInToGh(); err != nil {
			log.Fatal(err)
		}
	}

	if err := sc.ValidateSetUpConfig(); err != nil {
		return err
	}

	if AzAppExists(sc.AppName) {
		log.Fatal("App already exists")
	} else if err := sc.createAzApp(); err != nil {
		return err
	}

	if !sc.ServicePrincipalExists() {
		if err := sc.CreateServicePrincipal(); err != nil {
			return err
		}
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
		sc.createFederatedCredentials()
	}

	sc.setAzClientId()
	sc.setAzSubscriptionId()
	sc.setAzTenantId()

	log.Debug("Github connection with azure completed successfully!")
	return nil
}

func (sc *SetUpCmd) createAzApp() error {
	log.Debug("Commencing Azure app creation...")

	for {
		createAppCmd := exec.Command("az", "ad", "app", "create", "--only-show-errors", "--display-name", sc.AppName)

		out, err := createAppCmd.CombinedOutput()
		if err != nil {
			return err
		}

		log.Debug("Waiting 3 seconds to allow app time to populate")
		time.Sleep(3 * time.Second)

		if AzAppExists(sc.AppName) {
			var azApp map[string]interface{}
			json.Unmarshal(out, &azApp)
			appId := fmt.Sprint(azApp["appId"])

			sc.appId = appId

			log.Debug("App created successfully!")
			break
		}

	}

	return nil
}

func (sc *SetUpCmd) CreateServicePrincipal() error {
	log.Debug("Creating Azure service principal...")

	var exponentialBackOffCeilingSecs int64 = 1800 // 30 min
	lastUpdatedAt := time.Now()
	attempts := 0

	for attempts < 10 {
		if time.Since(lastUpdatedAt).Hours() >= 12 {
			attempts = 0
		}

		lastUpdatedAt = time.Now()
		attempts += 1

		delaySecs := int64(math.Floor((math.Pow(2, float64(attempts)) - 1) * 0.5))
		if delaySecs > exponentialBackOffCeilingSecs {
			delaySecs = exponentialBackOffCeilingSecs
		}

		createSpCmd := exec.Command("az", "ad", "sp", "create", "--id", sc.appId, "--only-show-errors")
		out, err := createSpCmd.CombinedOutput()
		if err != nil {
			log.Fatal(out)
			return err
		}

		log.Debug("Waiting 3 seconds to allow service principal time to populate")
		time.Sleep(time.Duration(delaySecs))

		if sc.ServicePrincipalExists() {
			var servicePrincipal map[string]interface{}
			json.Unmarshal(out, &servicePrincipal)
			objectId := fmt.Sprint(servicePrincipal["objectId"])

			sc.spObjectId = objectId
			log.Debug("Service principal created successfully!")
			break
		}
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
	json.Unmarshal(out, &tenantId)
	tenantId = fmt.Sprint(tenantId)

	sc.tenantId = tenantId

	return nil
}

func (sc *SetUpCmd) ValidateSetUpConfig() error {
	log.Debug("Checking that provided information is valid...")

	if !IsSubscriptionIdValid(sc.SubscriptionID) {
		return errors.New("subscription id is not valid")
	}

	if !isValidResourceGroup(sc.ResourceGroupName) {
		return errors.New("resource group is not valid")
	}

	if sc.AppName == "" {
		return errors.New("invalid app name")
	}

	if !isValidGhRepo(sc.Repo) {
		return errors.New("github repo is not valid")
	}

	return nil
}

func (sc *SetUpCmd) hasFederatedCredentials() bool {
	log.Debug("Checking for existing federated credentials...")
	uri := fmt.Sprintf("https://graph.microsoft.com/beta/applications/%s/federatedIdentityCredentials", sc.appObjectId)
	getFicCmd := exec.Command("az", "rest", "--method", "GET", "--uri", uri, "--query", "value")
	out, err := getFicCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var fics []interface{}
	json.Unmarshal(out, &fics)

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
	filter := fmt.Sprintf("displayName eq '%s'", sc.AppName)
	getObjectIdCmd := exec.Command("az", "ad", "app", "list", "--only-show-errors", "--filter", filter, "--query", "[].objectId")
	out, err := getObjectIdCmd.CombinedOutput()
	if err != nil {
		log.Fatalf(string(out))
		return err
	}

	var objectId []string
	json.Unmarshal(out, &objectId)
	objId := objectId[0]

	sc.appObjectId = objId

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
