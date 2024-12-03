package providers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
)

// EnsureAzCli ensures that the Azure CLI is installed and the user is logged in
func (az *AzClient) EnsureAzCli() error {
	if err := az.ValidateAzCliInstalled(); err != nil {
		return fmt.Errorf("failed to validate az CLI installation: %w", err)
	}

	if err := az.EnsureAzCliLoggedIn(); err != nil {
		return fmt.Errorf("failed to ensure az CLI login: %w", err)
	}

	return nil
}

func (az *AzClient) GetAzCliVersion() (string, error) {
	out, err := az.CommandRunner.RunCommand("az", "version", "-o", "json")
	if err != nil {
		return "", errors.New("unable to obtain az cli version")
	}

	var version map[string]interface{}
	if err := json.Unmarshal([]byte(out), &version); err != nil {
		return "", errors.New("unable to unmarshal az cli version output to map")
	}

	return fmt.Sprint(version["azure-cli"]), nil
}

func (az *AzClient) GetAzUpgrade() string {
	selection := &promptui.Select{
		Label: "Your Azure CLI version must be at least 2.37.0 - would you like us to update it for you?",
		Items: []string{"yes", "no"},
	}

	_, selectResponse, err := selection.Run()
	if err != nil {
		return err.Error()
	}

	return selectResponse
}

func (az *AzClient) UpgradeAzCli() {
	_, err := az.CommandRunner.RunCommand("az", "upgrade", "-y")
	if err != nil {
		log.Fatal("Error: unable to upgrade az cli version; ", err)
	}

	log.Info("Azure CLI upgrade was successful!")
}

func (az *AzClient) ValidateAzCliInstalled() error {
	log.Debug("Checking that Azure Cli is installed...")
	_, err := az.CommandRunner.RunCommand("az")
	if err != nil {
		return errors.New("az cli not installed. Find installation instructions at this link: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli")
	}
	azCliVersion, err := az.GetAzCliVersion()
	if err != nil {
		return fmt.Errorf("getting azcli version: %w", err)
	}
	currentVersion, err := version.NewVersion(azCliVersion)
	if err != nil {
		return fmt.Errorf("parsing azcli version: %w", err)
	}

	constraints, err := version.NewConstraint(">= 2.37")
	if err != nil {
		return fmt.Errorf("getting azcli version constraint: %w", err)
	}

	if !constraints.Check(currentVersion) {
		if ans := az.GetAzUpgrade(); ans == "no" {
			return fmt.Errorf("az cli version must be at least 2.37.0, but current version is %s", azCliVersion)
		}
		az.UpgradeAzCli()
	}
	return nil
}

func (az *AzClient) IsLoggedInToAz() bool {
	log.Debug("Checking that user is logged in to Azure CLI...")
	_, err := az.CommandRunner.RunCommand("az", "ad", "signed-in-user", "show", "--only-show-errors", "--query", "objectId")
	return err != nil
}

func (az *AzClient) EnsureAzCliLoggedIn() error {
	if !az.IsLoggedInToAz() {
		if err := az.LogInToAz(); err != nil {
			return fmt.Errorf("unable to log in to Azure: %w", err)
		}
	}
	return nil
}

func (az *AzClient) LogInToAz() error {
	log.Debug("Logging user in to Azure Cli...")
	_, err := az.CommandRunner.RunCommand("az", "login", "--allow-no-subscriptions")
	if err != nil {
		return err
	}

	log.Debug("Successfully logged in!")
	return nil
}

func (az *AzClient) IsSubscriptionIdValid(subscriptionId string) error {
	if subscriptionId == "" {
		return errors.New("subscriptionId cannot be empty")
	}

	out, err := az.CommandRunner.RunCommand("az", "account", "show", "-s", subscriptionId, "--query", "id")
	if err != nil {
		return err
	}

	var azSubscription string
	if err = json.Unmarshal([]byte(out), &azSubscription); err != nil {
		return err
	}

	if azSubscription == "" {
		return errors.New("subscription not found")
	}

	return nil
}

func (az *AzClient) IsValidResourceGroup(
	subscriptionId string,
	resourceGroup string,
) error {
	if resourceGroup == "" {
		return errors.New("resource group cannot be empty")
	}

	query := fmt.Sprintf("[?name=='%s']", resourceGroup)
	out, err := az.CommandRunner.RunCommand("az", "group", "list", "--subscription", subscriptionId, "--query", query)
	if err != nil {
		log.Errorf("failed to validate resource group %q from subscription %q: %s", resourceGroup, subscriptionId, err)
		return err
	}

	var rg []interface{}
	if err = json.Unmarshal([]byte(out), &rg); err != nil {
		return err
	}

	if len(rg) == 0 {
		return fmt.Errorf("resource group %q not found from subscription %q", resourceGroup, subscriptionId)
	}

	return nil
}

func (az *AzClient) AzAppExists(appName string) bool {
	log.Debugf("Checking if app %q exists...", appName)
	filter := fmt.Sprintf("displayName eq '%s'", appName)
	out, err := az.CommandRunner.RunCommand("az", "ad", "app", "list", "--only-show-errors", "--filter", filter, "--query", "[].appId")
	if err != nil {
		return false
	}

	var azApp []string
	json.Unmarshal([]byte(out), &azApp)

	return len(azApp) >= 1
}

func (az *AzClient) GetServicePrincipal(appId string) (string, error) {
	out, err := az.CommandRunner.RunCommand("az", "ad", "sp", "show", "--only-show-errors", "--id", appId, "--query", "id")
	if err != nil {
		return "", err
	}

	var objectId string
	json.Unmarshal([]byte(out), &objectId)

	log.Debugf("Service principal with appId '%s' exists", appId)
	return objectId, nil
}

func (az *AzClient) AzAcrExists(acrName string) bool {
	query := fmt.Sprintf("[?name=='%s']", acrName)
	out, err := az.CommandRunner.RunCommand("az", "acr", "list", "--only-show-errors", "--query", query)
	if err != nil {
		return false
	}

	var azAcr []interface{}
	json.Unmarshal([]byte(out), &azAcr)

	return len(azAcr) >= 1
}

func (az *AzClient) AzAksExists(aksName string, resourceGroup string) bool {
	_, err := az.CommandRunner.RunCommand("az", "aks", "browse", "-g", resourceGroup, "--name", aksName)
	return err == nil
}

func (az *AzClient) GetCurrentAzSubscriptionLabel() (SubLabel, error) {
	out, err := az.CommandRunner.RunCommand("az", "account", "show", "--query", "{id: id, name: name}")
	if err != nil {
		log.Fatal(err)
	}

	var currentSub SubLabel
	if err := json.Unmarshal([]byte(out), &currentSub); err != nil {
		return SubLabel{}, fmt.Errorf("failed to unmarshal JSON output: %v", err)
	} else if currentSub.ID == "" {
		return SubLabel{}, errors.New("no current subscription found")
	}

	return currentSub, nil
}

func (az *AzClient) GetAzSubscriptionLabels() ([]SubLabel, error) {
	out, err := az.CommandRunner.RunCommand("az", "account", "list", "--all", "--query", "[].{id: id, name: name}")
	if err != nil {
		log.Fatal(err)
	}

	var subLabels []SubLabel
	if err := json.Unmarshal([]byte(out), &subLabels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON output: %v", err)
	} else if len(subLabels) == 0 {
		return nil, errors.New("no subscriptions found")
	}

	return subLabels, nil
}
