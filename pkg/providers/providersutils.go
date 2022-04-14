package providers

import (
	"fmt"
	"encoding/json"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func CheckAzCliInstalled()  {
	log.Debug("Checking that Azure Cli is installed...")
	azCmd := exec.Command("az")
	_, err := azCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error: AZ cli not installed. Find installation instructions at this link: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli")
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
	azCmd.Stdout  = os.Stdout
	azCmd.Stderr = os.Stderr
	err := azCmd.Run()
	if err != nil {
		return err
	}

	log.Debug("Successfully logged in!")
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

func isValidResourceGroup(resourceGroup string) bool {
	if resourceGroup == "" {
		return false
	}

	query := fmt.Sprintf("[?name=='%s']", resourceGroup)
	getResourceGroupCmd := exec.Command("az", "group", "list", "--query", query)
	out, err := getResourceGroupCmd.CombinedOutput()
	if err != nil {
		return false
	}

	var rg []interface{}
	json.Unmarshal(out, &rg)

	if len(rg) ==  0 {
		return false
	}

	return true
}

func isValidGhRepo(repo string) bool {
	listReposCmd := exec.Command("gh", "repo", "view", repo)
		_, err := listReposCmd.CombinedOutput()
		if err != nil {
			log.Fatal("Github repo not found")
			return false
		}
		return true
}

func AzAppExists(appName string) bool {
	filter := fmt.Sprintf("displayName eq '%s'", appName)
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

func (sc *SetUpCmd) ServicePrincipalExists() bool {
	filter := fmt.Sprintf("appId eq '%s'", sc.appId)
	checkSpExistsCmd := exec.Command("az", "ad", "sp","list", "--only-show-errors", "--filter", filter, "--query", "[].objectId")
	out, err := checkSpExistsCmd.CombinedOutput()
	if err != nil {
		return true
	}

	var azSp []string
	json.Unmarshal(out, &azSp)
	
	if len(azSp) == 1 {
		log.Debug("Service principal already exists - skipping service principal creation.")
		// TODO: tell user sp already exists and ask if they want to use it?
		objectId := fmt.Sprint(azSp[0])
		sc.spObjectId = objectId
		return true
	}

	return false
}

func AzAcrExists(acrName string) bool {
	query := fmt.Sprintf("[?name=='%s']", acrName)
	checkAcrExistsCmd := exec.Command("az", "acr","list", "--only-show-errors", "--query", query)
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
	checkAksExistsCmd := exec.Command("az", "aks","browse", "-g", resourceGroup, "--name", aksName)
	_, err := checkAksExistsCmd.CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

