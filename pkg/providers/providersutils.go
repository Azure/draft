package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func CheckAzCliInstalled() error {
	log.Debug("Checking that Azure CLI is installed...")
	azCmd := exec.Command("az")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		return errors.New("Error: Azure CLI is not installed. Find installation instructions at this link: https://docs.microsoft.com/cli/azure/install-azure-cli")
	}
	return nil
}

func IsLoggedInToAz() (bool, error) {
	log.Debug("Checking that user is logged in to Azure CLI...")
	azCmd := exec.Command("az", "ad", "signed-in-user", "show", "--only-show-errors", "--query", "objectId")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return true, nil
}

func HasGhCli() bool {
	log.Debug("Checking that GitHub CLI is installed...")
	ghCmd := exec.Command("gh")
	_, err := ghCmd.CombinedOutput()
	if err != nil {
		log.Println("Error: The GitHub CLI is required to complete this process. Find installation instructions at this link: https://cli.github.com/manual/installation")
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
		log.Printf("%s\n", out)
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

func isValidGhRepo(repo string) (bool, error) {
	listReposCmd := exec.Command("gh", "repo", "view", repo)
	_, err := listReposCmd.CombinedOutput()
	if err != nil {
		log.Println("GitHub repo not found")
		return false, err
	}
	return true, nil
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

func GetCurrentAzSubscriptionId() ([]string, error) {
	err := CheckAzCliInstalled()
	if err != nil {
		return nil, &json.InvalidUnmarshalError{}
	}
	isLoggedIn, err := IsLoggedInToAz()
	if err != nil {
		return nil, err
	}
	if !isLoggedIn {
		if err := LogInToAz(); err != nil {
			return nil, err
		}
	}

	getAccountCmd := exec.Command("az", "account", "show", "--query", "[id]")
	out, err := getAccountCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var ids []string
	err = json.Unmarshal(out, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
