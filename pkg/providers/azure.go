package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
)

type SetUpCmd struct {
	AppName           string
	SubscriptionID    string
	ResourceGroupName string
	Provider          string
	appId string
	tenantId string
	clientId string
	objectId string
}

func InitiateAzureOIDCFlow(sc *SetUpCmd) error {
	checkAzCliInstalled()

	if err := sc.ValidateSetUpConfig(); err != nil {
		return err
	}

	if !sc.appExistsAlready() {
		if err := sc.createAzApp(); err != nil {
			return err
		}
	} 

	if !sc.serviceProviderExistsAlready() {
		if err := sc.CreateServiceProvider(); err != nil {
			return err
		}
	}

	if err := sc.assignSpRole(); err != nil {
		return err
	}

	return nil
}

func checkAzCliInstalled()  {
	azCmd := exec.Command("az")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
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
	
	if len(azApp) == 1 {
		// TODO: tell user app already exists and ask if they want to use it?
		appId := fmt.Sprint(azApp[0])
		sc.appId = appId
		return true
	}

	return false
}

func (sc *SetUpCmd) createAzApp() error {
	// TODO: pull only appId info from az app create response
	// createAppCmd := exec.Command("az", "ad", "app", "create", "--only-show-errors", "--display-name", sc.appName)

	// using the az show app command for testing purposes
	createAppCmd := exec.Command("az", "ad", "app", "show", "--id", "864b58c9-1c86-4e22-a472-f866438378d0")
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
	createSpCmd := exec.Command("az", "ad", "sp", "create", "--id", sc.appId)
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
	assignSpRoleCmd := exec.Command("az", "role", "assignment", "create", "--role", "contributor", "--subscription", sc.SubscriptionID, "--assignee-object-id", sc.objectId, "--assignee-principle-type", "ServicePrincipal", "--scope", scope)
	out, err := assignSpRoleCmd.CombinedOutput()
	if err != nil {
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
	// TODO: check subscriptionID is valid

	if sc.AppName == "" {
		return errors.New("Invalid app name")
	} else if sc.ResourceGroupName == "" {
		return errors.New("Invalid resource group name")
	}

	return nil
}
