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

	if err := sc.CreateServiceProvider(); err != nil {
		return err
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

func (sc *SetUpCmd) CreateServiceProvider() error {
	// TODO: set context to correct subscription
	// if err := sc.setAZContext(); err != nil {
	// 	return err
	// }

	// createAppCmd := exec.Command("az", "ad", "app", "create", "--only-show-errors", "--display-name", sc.appName)
	// using the az show app command for testing purposes
	createAppCmd := exec.Command("az", "ad", "app", "show", "--id", "864b58c9-1c86-4e22-a472-f866438378d0")
	stdoutStderr, err := createAppCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", stdoutStderr)
		return err
	}

	var azApp map[string]interface{}
	json.Unmarshal(stdoutStderr, &azApp)
	appId := fmt.Sprint(azApp["appId"])

	fmt.Println(appId)

	createSPCmd := exec.Command("az", "ad", "sp", "create", "--id", appId)
	out, sperr := createSPCmd.CombinedOutput()
	if sperr != nil {
		return sperr
	}

	var serviceProvider map[string]interface{}
	json.Unmarshal(out, &serviceProvider)
	objectId := fmt.Sprint(serviceProvider["objectId"])

	fmt.Println(objectId)
	return nil
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
