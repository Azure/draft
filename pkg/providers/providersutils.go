package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/hashicorp/go-version"
	"github.com/manifoldco/promptui"
)

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

func CheckAzCliInstalled() {
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

func HasGhCli() bool {
	log.Debug("Checking that github cli is installed...")
	ghCmd := exec.Command("gh")
	_, err := ghCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error: The github cli is required to complete this process. Find installation instructions at this link: https://cli.github.com/manual/installation")
		return false
	}

	log.Debug("Github cli found!")
	return true
}

func IsLoggedInToGh() bool {
	log.Debug("Checking that user is logged in to github...")
	ghCmd := exec.Command("gh", "auth", "status")
	out, err := ghCmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		return false
	}

	log.Debug("User is logged in!")
	return true

}

func LogInToGh() error {
	log.Debug("Logging user in to github...")
	ghCmd := exec.Command("gh", "auth", "login")
	ghCmd.Stdin = os.Stdin
	ghCmd.Stdout = os.Stdout
	ghCmd.Stderr = os.Stderr
	err := ghCmd.Run()
	if err != nil {
		return err
	}

	return nil
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

func isValidResourceGroup(resourceGroup string) error {
	if resourceGroup == "" {
		return errors.New("resource group cannot be empty")
	}

	query := fmt.Sprintf("[?name=='%s']", resourceGroup)
	getResourceGroupCmd := exec.Command("az", "group", "list", "--query", query)
	out, err := getResourceGroupCmd.CombinedOutput()
	if err != nil {
		log.Errorf("failed to validate resourcegroup: %s", err)
		return err
	}

	var rg []interface{}
	if err = json.Unmarshal(out, &rg); err != nil {
		return err
	}

	if len(rg) == 0 {
		return errors.New("resource group not found")
	}

	return nil
}

func isValidGhRepo(repo string) error {
	listReposCmd := exec.Command("gh", "repo", "view", repo)
	_, err := listReposCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Github repo not found")
		return err
	}
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

func GetCurrentAzSubscriptionId() []string {
	CheckAzCliInstalled()
	if !IsLoggedInToAz() {
		if err := LogInToAz(); err != nil {
			log.Fatal(err)
		}
	}

	getAccountCmd := exec.Command("az", "account", "show", "--query", "[id]")
	out, err := getAccountCmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	var ids []string
	json.Unmarshal(out, &ids)

	return ids
}
