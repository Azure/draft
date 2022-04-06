package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
	

	"github.com/Azure/draftv2/pkg/osutil"
	log "github.com/sirupsen/logrus"
)

type SetUpCmd struct {
	AppName           string
	SubscriptionID    string
	ResourceGroupName string
	Provider          string
	Repo string
	appId string
	tenantId string
	appObjectId string
	spObjectId string
}

type federatedIdentityCredentials struct {
	Name string `json:"name"`
	Issuer string `json:"issuer"`
	Subject string `json:"subject"`
	Description string `json:"description"`
	Audiences []string 	`json:"audiences"`
}

func InitiateAzureOIDCFlow(sc *SetUpCmd) error {
	if !osutil.HasGhCli() || !osutil.IsLoggedInToGh() {
		log.Fatal("Error: Unable to login to your github account.")
	}

	if err := sc.ValidateSetUpConfig(); err != nil {
		return err
	}


	if sc.appExistsAlready() {
		log.Fatal("App already exists")
	} else if err := sc.createAzApp(); err != nil {
		return err
	}

	if !sc.servicePrincipalExistsAlready() {
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

	return nil
}


func (sc *SetUpCmd) appExistsAlready() bool {
	filter := fmt.Sprintf("displayName eq '%s'", sc.AppName)
	checkAppExistsCmd := exec.Command("az", "ad", "app","list", "--only-show-errors", "--filter", filter, "--query", "[].appId")
	out, err := checkAppExistsCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var azApp []string
	json.Unmarshal(out, &azApp)
	
	if len(azApp) >= 1 {
		// TODO: tell user app already exists and ask which one they want to use?
		return true
	}

	return false
}

func (sc *SetUpCmd) createAzApp() error {
	// TODO: need to change command to force create app? or ask for new app name?
	createAppCmd := exec.Command("az", "ad", "app", "create", "--only-show-errors", "--display-name", sc.AppName)

	out, err := createAppCmd.CombinedOutput()
	if err != nil {
		return err
	}

	var azApp map[string]interface{}
	json.Unmarshal(out, &azApp)
	appId := fmt.Sprint(azApp["appId"])

	sc.appId = appId
	return nil
}

func (sc *SetUpCmd) servicePrincipalExistsAlready() bool {
	filter := fmt.Sprintf("appId eq '%s'", sc.appId)
	checkSpExistsCmd := exec.Command("az", "ad", "sp","list", "--only-show-errors", "--filter", filter, "--query", "[].objectId")
	out, err := checkSpExistsCmd.CombinedOutput()
	if err != nil {
		return true
	}

	var azSp []string
	json.Unmarshal(out, &azSp)
	
	if len(azSp) == 1 {
		// TODO: tell user sp already exists and ask if they want to use it?
		objectId := fmt.Sprint(azSp[0])
		sc.spObjectId = objectId
		return true
	}

	return false
}

func (sc *SetUpCmd) CreateServicePrincipal() error {
	createSpCmd := exec.Command("az", "ad", "sp", "create", "--id", sc.appId, "--only-show-errors")
	out, err := createSpCmd.CombinedOutput()
	if err != nil {
		log.Fatal(out)
		return err
	}

	var servicePrincipal map[string]interface{}
	json.Unmarshal(out, &servicePrincipal)
	objectId := fmt.Sprint(servicePrincipal["objectId"])

	sc.spObjectId = objectId

	return nil
}

func (sc *SetUpCmd) assignSpRole() error {
	scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", sc.SubscriptionID, sc.ResourceGroupName)
	assignSpRoleCmd := exec.Command("az", "role", "assignment", "create", "--role", "contributor", "--subscription", sc.SubscriptionID, "--assignee-object-id", sc.spObjectId, "--assignee-principal-type", "ServicePrincipal", "--scope", scope, "--only-show-errors")
	out, err := assignSpRoleCmd.CombinedOutput()
	if err != nil {
		log.Fatalf(string(out))
		return err
	}

	return nil
}

func (sc *SetUpCmd) getTenantId() error {
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
	if !IsSubscriptionIdValid(sc.SubscriptionID) {
		return errors.New("Subscription id is not valid")
	}

	if sc.AppName == "" {
		return errors.New("Invalid app name")
	} else if sc.ResourceGroupName == "" {
		return errors.New("Invalid resource group name")
	}

	if !sc.ValidGhRepo() {
		return errors.New("Github repo is not valid")
	}

	return nil
}

func IsSubscriptionIdValid(subscriptionId string) bool {
	if subscriptionId == "" { 
		return false
	}

	getSubscriptionIdCmd := exec.Command("az", "account", "show", "-s", subscriptionId, "--query", "id")
	out, err := getSubscriptionIdCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var azSubscription string
	json.Unmarshal(out, &azSubscription)

	if azSubscription != "" {
		return true
	}

	return false
}

func (sc *SetUpCmd) hasFederatedCredentials() bool {
	uri := fmt.Sprintf("https://graph.microsoft.com/beta/applications/%s/federatedIdentityCredentials", sc.appObjectId)
	getFicCmd := exec.Command("az", "rest", "--method", "GET", "--uri", uri, "--query", "value")
	out, err := getFicCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var fics []interface{}
	json.Unmarshal(out, &fics)

	if len(fics) > 0 {
		// TODO: ask user if they want to use current credentials?
		// TODO: check if fics with the name we want exist already
		return true
	}

	return false
}

func (sc *SetUpCmd) ValidGhRepo() bool {
	listReposCmd := exec.Command("gh", "repo", "view", sc.Repo)
		_, err := listReposCmd.CombinedOutput()
		if err != nil {
			log.Fatal("Github repo not found")
			return false
		}
		return true
}


func (sc *SetUpCmd) createFederatedCredentials() error {
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

	time.Sleep(10 * time.Second)
	count := 0

	// check to make sure credentials were created
	// count to prevent infinite loop
	for count < 10	{
		if sc.hasFederatedCredentials() {
			break
		}

		log.Info("Credentials not yet created, retrying...")
		count += 1
	}

	return nil

}

func (sc *SetUpCmd) getAppObjectId() error {
	filter := fmt.Sprintf("displayName eq '%s'", sc.AppName)
	getObjectIdCmd := exec.Command("az", "ad", "app","list", "--only-show-errors", "--filter", filter, "--query", "[].objectId")
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

