package providers

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type SubLabel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EnsureGhCliInstalled ensures that the Github CLI is installed and the user is logged in
func EnsureGhCli() {
	EnsureGhCliInstalled()
	EnsureGhCliLoggedIn()
}

func EnsureGhCliInstalled() {
	log.Debug("Checking that github cli is installed...")
	ghCmd := exec.Command("gh")
	_, err := ghCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Error: The github cli is required to complete this process. Find installation instructions at this link: https://github.com/cli/cli#installation")
	}

	log.Debug("Github cli found!")
}

func EnsureGhCliLoggedIn() {
	EnsureGhCliInstalled()
	if !IsLoggedInToGh() {
		if err := LogInToGh(); err != nil {
			log.Fatal("Error: unable to log in to github")
		}
	}
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

func isValidGhRepo(repo string) error {
	listReposCmd := exec.Command("gh", "repo", "view", repo)
	_, err := listReposCmd.CombinedOutput()
	if err != nil {
		log.Fatal("Github repo not found")
		return err
	}
	return nil
}
