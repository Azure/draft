package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"

	
	"github.com/Azure/draftv2/pkg/osutil"
)

type SetUpCmd struct {
	AppName           string
	SubscriptionID    string
	ResourceGroupName string
	Provider          string
	Repo string
	appId string
	tenantId string
	clientId string
	objectId string
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
	
	if !sc.serviceProviderExistsAlready() {
		if err := sc.CreateServiceProvider(); err != nil {
			return err
		}
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

	// using the az show app command for testing purposes
	// createAppCmd := exec.Command("az", "ad", "app", "show", "--id", "864b58c9-1c86-4e22-a472-f866438378d0")
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

func (sc *SetUpCmd) serviceProviderExistsAlready() bool {
	filter := fmt.Sprintf("appId eq '%s'", sc.appId)
	checkSpExistsCmd := exec.Command("az", "ad", "sp","list", "--only-show-errors", "--filter", filter, "--query", "[].objectId")
	out, err := checkSpExistsCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var azSp []string
	json.Unmarshal(out, &azSp)
	
	if len(azSp) == 1 {
		// TODO: tell user sp already exists and ask if they want to use it?
		objectId := fmt.Sprint(azSp[0])
		sc.objectId = objectId
		return true
	}

	return false
}

func (sc *SetUpCmd) CreateServiceProvider() error {
	createSpCmd := exec.Command("az", "ad", "sp", "create", "--id", sc.appId, "--only-show-errors")
	out, err := createSpCmd.CombinedOutput()
	if err != nil {
		return err
	}

	var serviceProvider map[string]interface{}
	json.Unmarshal(out, &serviceProvider)
	objectId := fmt.Sprint(serviceProvider["objectId"])

	sc.objectId = objectId

	return nil
}

func (sc *SetUpCmd) assignSpRole() error {
	scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", sc.SubscriptionID, sc.ResourceGroupName)
	assignSpRoleCmd := exec.Command("az", "role", "assignment", "create", "--role", "contributor", "--subscription", sc.SubscriptionID, "--assignee-object-id", sc.objectId, "--assignee-principal-type", "ServicePrincipal", "--scope", scope, "--only-show-errors")
	out, err := assignSpRoleCmd.CombinedOutput()
	if err != nil {
		log.Fatalf(string(out))
		return err
	}

	var serviceProvider map[string]interface{}
	json.Unmarshal(out, &serviceProvider)
	clientId := fmt.Sprint(serviceProvider["clientId"])
	tenantId := fmt.Sprint(serviceProvider["tenantId"])

	sc.clientId = clientId
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
	uri := fmt.Sprintf("https://graph.microsoft.com/beta/applications/%s/federatedIdentityCredentials", sc.objectId)
	getFicCmd := exec.Command("az", "rest", "--method", "GET", "--uri", uri, "--query", "value")
	out, err := getFicCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var fics map[string]interface{}
	json.Unmarshal(out, &fics)

	if len(fics) > 0 {
		// TODO: ask user if they want to use current credentials?
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
	fics := []federatedIdentityCredentials{
		{Name: "prfic", Subject: "repo:%s:pull_request", Issuer: "https://token.actions.githubusercontent.com", Description: "pr", Audiences: []string{"api://AzureADTokenExchange"}},
		{Name: "mainfic", Subject: "repo:%s:ref:refs/heads/main", Issuer: "https://token.actions.githubusercontent.com", Description: "main", Audiences: []string{"api://AzureADTokenExchange"}},
		{Name: "masterfic", Subject: "repo:%s:ref:refs/heads/master", Issuer: "https://token.actions.githubusercontent.com", Description: "master", Audiences: []string{"api://AzureADTokenExchange"}},
	}

	uri := fmt.Sprintf("https://graph.microsoft.com/beta/applications/%s/federatedIdentityCredentials", sc.appId)

	for _, fic := range fics {
		subject := fmt.Sprintf(fic.Subject, sc.Repo)
		fic.Subject = subject

		ficBody, err := json.Marshal(fic)
		if err != nil {
			return err
		}

		createFicCmd := exec.Command("az", "rest", "--method", "POST", "--uri", uri, "--body", string(ficBody))
		out, ficErr := createFicCmd.CombinedOutput()
		if ficErr != nil {
			log.Fatalf(string(out))
			return ficErr
		}

	}

	time.Sleep(5 * time.Second)

	// check to make sure credentials were created
	sc.hasFederatedCredentials()	

	return nil

}

