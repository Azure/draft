package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
)

type SetUpCmd struct {
	AppName           string
	SubscriptionID    string
	ResourceGroupName string
	Provider          string
}

func InitiateAzureOIDCFlow(sc *SetUpCmd) error {
	if err := sc.ValidateSetUpConfig(); err != nil {
		return err
	}

	// TODO: set context to correct subscription
	// if err := sc.setAZContext(); err != nil {
	// 	return err
	// }

	appId, appErr := createAzApp()
	if appErr != nil {
		fmt.Println(appId)
		return appErr
	}

	clientId, tenantId, spErr := sc.CreateServiceProvider(appId)
	if spErr != nil {
		fmt.Println(clientId, tenantId)
		return spErr
	}


	return nil
}

func (sc *SetUpCmd) setAZContext() error {
	setContextCmd := exec.Command("az", "account", "set", "--subscription", sc.SubscriptionID)
	stdoutStderr, err := setContextCmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", stdoutStderr)

	return nil
}

func createAzApp() (string, error) {
	// TODO: check if app already exists,if not run this logic

	// createAppCmd := exec.Command("az", "ad", "app", "create", "--only-show-errors", "--display-name", sc.appName)
	// using the az show app command for testing purposes
	createAppCmd := exec.Command("az", "ad", "app", "show", "--id", "864b58c9-1c86-4e22-a472-f866438378d0")
	out, err := createAppCmd.CombinedOutput()
	if err != nil {
		errMessage := string(out)
		return errMessage, err
	}

	var azApp map[string]interface{}
	json.Unmarshal(out, &azApp)
	appId := fmt.Sprint(azApp["appId"])

	fmt.Println(appId)
	return appId, nil
}

func (sc *SetUpCmd) CreateServiceProvider(appId string) (string,  string, error) {
	createSpCmd := exec.Command("az", "ad", "sp", "create", "--id", appId)
	spOut, spErr := createSpCmd.CombinedOutput()
	if spErr != nil {
		return "create sp failed\t", string(spOut), spErr
	}

	var serviceProvider map[string]interface{}
	json.Unmarshal(spOut, &serviceProvider)
	objectId := fmt.Sprint(serviceProvider["objectId"])

	scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", sc.SubscriptionID, sc.ResourceGroupName)
	assignSpRoleCmd := exec.Command("az", "role", "assignment", "create", "--role", "contributor", "--subscription", sc.SubscriptionID, "--assignee-object-id", objectId, "--assignee-principle-type", "ServicePrincipal", "--scope", scope)
	roleOut, roleErr := assignSpRoleCmd.CombinedOutput()
	if roleErr != nil {
		return "assign sp role failed\t", string(roleOut), roleErr
	}

	json.Unmarshal(roleOut, &serviceProvider)
	clientId := fmt.Sprint(serviceProvider["clientId"])
	tenantId := fmt.Sprint(serviceProvider["tenantId"])

	return clientId, tenantId, nil
}

func (sc *SetUpCmd) ValidateSetUpConfig() error {
	//fmt.Printf("%v", sc)

	// TODO: check subscriptionID length
	_, err := strconv.ParseFloat(sc.SubscriptionID, 64)
	if err != nil {
		return errors.New("Invalid number")
	}

	if sc.AppName == "" {
		return errors.New("Invalid app name")
	} else if sc.ResourceGroupName == "" {
		return errors.New("Invalid resource group name")
	}

	return nil
}
