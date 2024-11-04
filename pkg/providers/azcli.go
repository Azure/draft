package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-version"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
)

// EnsureAzCli ensures that the Azure CLI is installed and the user is logged in
func EnsureAzCli() {
	EnsureAzCliInstalled()
	EnsureAzCliLoggedIn()
}

func GetAzCliVersion() string {
	azCmd := exec.Command("az", "version", "-o", "json")
	out, err := azCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error: unable to obtain az cli version")
	}

	var version map[string]interface{}
	if err := json.Unmarshal(out, &version); err != nil {
		log.Fatal("unable to unmarshal az cli version output to map")
	}

	return fmt.Sprint(version["azure-cli"])
}

func getAzUpgrade() string {
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

func upgradeAzCli() {
	azCmd := exec.Command("az", "upgrade", "-y")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error: unable to upgrade az cli version; ", err)
	}

	log.Info("Azure CLI upgrade was successful!")
}

func EnsureAzCliInstalled() {
	log.Debug("Checking that Azure Cli is installed...")
	azCmd := exec.Command("az")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error: AZ cli not installed. Find installation instructions at this link: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli")
	}

	currentVersion, err := version.NewVersion(GetAzCliVersion())
	if err != nil {
		log.Fatal(err)
	}

	constraints, err := version.NewConstraint(">= 2.37")
	if err != nil {
		log.Fatal(err)
	}

	if !constraints.Check(currentVersion) {
		if ans := getAzUpgrade(); ans == "no" {
			log.Fatal("Az cli version must be at least 2.37.0")
		}
		upgradeAzCli()
	}
}

func IsLoggedInToAz() bool {
	log.Debug("Checking that user is logged in to Azure CLI...")
	azCmd := exec.Command("az", "ad", "signed-in-user", "show", "--only-show-errors", "--query", "objectId")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

func EnsureAzCliLoggedIn() {
	EnsureAzCliInstalled()
	if !IsLoggedInToAz() {
		if err := LogInToAz(); err != nil {
			log.Fatal("Error: unable to log in to Azure")
		}
	}
}

func LogInToAz() error {
	log.Debug("Logging user in to Azure Cli...")
	azCmd := exec.Command("az", "login", "--allow-no-subscriptions")
	azCmd.Stdin = os.Stdin
	azCmd.Stdout = os.Stdout
	azCmd.Stderr = os.Stderr
	err := azCmd.Run()
	if err != nil {
		return err
	}

	log.Debug("Successfully logged in!")
	return nil
}

func AzAppExists(appName string) bool {
	filter := fmt.Sprintf("displayName eq '%s'", appName)
	checkAppExistsCmd := exec.Command("az", "ad", "app", "list", "--only-show-errors", "--filter", filter, "--query", "[].appId")
	out, err := checkAppExistsCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var azApp []string
	json.Unmarshal(out, &azApp)

	return len(azApp) >= 1
}

func (sc *SetUpCmd) ServicePrincipalExists() bool {
	checkSpExistsCmd := exec.Command("az", "ad", "sp", "show", "--only-show-errors", "--id", sc.appId, "--query", "id")
	out, err := checkSpExistsCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var objectId string
	json.Unmarshal(out, &objectId)

	log.Debug("Service principal exists")
	// TODO: tell user sp already exists and ask if they want to use it?
	sc.spObjectId = objectId
	return true
}

func AzAcrExists(acrName string) bool {
	query := fmt.Sprintf("[?name=='%s']", acrName)
	checkAcrExistsCmd := exec.Command("az", "acr", "list", "--only-show-errors", "--query", query)
	out, err := checkAcrExistsCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var azAcr []interface{}
	json.Unmarshal(out, &azAcr)

	if len(azAcr) >= 1 {
		return true
	}

	return false
}

func AzAksExists(aksName string, resourceGroup string) bool {
	checkAksExistsCmd := exec.Command("az", "aks", "browse", "-g", resourceGroup, "--name", aksName)
	_, err := checkAksExistsCmd.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

func GetCurrentAzSubscriptionLabel() (SubLabel, error) {
	EnsureAzCliInstalled()
	if !IsLoggedInToAz() {
		if err := LogInToAz(); err != nil {
			return SubLabel{}, fmt.Errorf("failed to log in to Azure CLI: %v", err)
		}
	}

	getAccountCmd := exec.Command("az", "account", "show", "--query", "{id: id, name: name}")
	out, err := getAccountCmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	var currentSub SubLabel
	if err := json.Unmarshal(out, &currentSub); err != nil {
		return SubLabel{}, fmt.Errorf("failed to unmarshal JSON output: %v", err)
	} else if currentSub.ID == "" {
		return SubLabel{}, errors.New("no current subscription found")
	}

	return currentSub, nil
}

func GetAzSubscriptionLabels() ([]SubLabel, error) {
	EnsureAzCliInstalled()
	if !IsLoggedInToAz() {
		if err := LogInToAz(); err != nil {
			return nil, fmt.Errorf("failed to log in to Azure CLI: %v", err)
		}
	}

	getAccountCmd := exec.Command("az", "account", "list", "--all", "--query", "[].{id: id, name: name}")

	out, err := getAccountCmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	var subLabels []SubLabel
	if err := json.Unmarshal(out, &subLabels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON output: %v", err)
	} else if len(subLabels) == 0 {
		return nil, errors.New("no subscriptions found")
	}

	return subLabels, nil
}

func IsSubscriptionIdValid(subscriptionId string) error {
	if subscriptionId == "" {
		return errors.New("subscriptionId cannot be empty")
	}

	getSubscriptionIdCmd := exec.Command("az", "account", "show", "-s", subscriptionId, "--query", "id")
	out, err := getSubscriptionIdCmd.CombinedOutput()
	if err != nil {
		return err
	}

	var azSubscription string
	if err = json.Unmarshal(out, &azSubscription); err != nil {
		return err
	}

	if azSubscription == "" {
		return errors.New("subscription not found")
	}

	return nil
}

func isValidResourceGroup(
	subscriptionId string,
	resourceGroup string,
) error {
	if resourceGroup == "" {
		return errors.New("resource group cannot be empty")
	}

	query := fmt.Sprintf("[?name=='%s']", resourceGroup)
	getResourceGroupCmd := exec.Command("az", "group", "list", "--subscription", subscriptionId, "--query", query)
	out, err := getResourceGroupCmd.CombinedOutput()
	if err != nil {
		log.Errorf("failed to validate resource group %q from subscription %q: %s", resourceGroup, subscriptionId, err)
		return err
	}

	var rg []interface{}
	if err = json.Unmarshal(out, &rg); err != nil {
		return err
	}

	if len(rg) == 0 {
		return fmt.Errorf("resource group %q not found from subscription %q", resourceGroup, subscriptionId)
	}

	return nil
}
