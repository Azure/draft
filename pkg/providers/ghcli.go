package providers

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type SubLabel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EnsureGhCliInstalled ensures that the Github CLI is installed and the user is logged in
func (gh GhCliClient) EnsureGhCli() {
	gh.EnsureGhCliInstalled()
	gh.EnsureGhCliLoggedIn()
}

type GhClient interface {
	EnsureGhCli()
	EnsureGhCliLoggedIn()
	IsLoggedInToGh() bool
	LogInToGh() error
	IsValidGhRepo(repo string) error
	GetRepoNameWithOwner() (string, error)
}

var _ GhClient = &GhCliClient{}

type GhCliClient struct {
	CommandRunner CommandRunner
}

func NewGhClient() *GhCliClient {
	gh := &GhCliClient{
		CommandRunner: &DefaultCommandRunner{},
	}
	gh.EnsureGhCli()
	return gh
}

func (gh GhCliClient) exec(args ...string) (string, error) {
	return gh.CommandRunner.RunCommand(args...)
}

func (gh GhCliClient) EnsureGhCliInstalled() {
	log.Debug("Checking that github cli is installed...")
	_, err := gh.exec("gh")
	if err != nil {
		log.Fatal("Error: The github cli is required to complete this process. Find installation instructions at this link: https://github.com/cli/cli#installation")
	}

	log.Debug("Github cli found!")
}

func (gh GhCliClient) EnsureGhCliLoggedIn() {
	gh.EnsureGhCliInstalled()
	if !gh.IsLoggedInToGh() {
		if err := gh.LogInToGh(); err != nil {
			log.Fatal("Error: unable to log in to github")
		}
	}
}

func (gh GhCliClient) IsLoggedInToGh() bool {
	log.Debug("Checking that user is logged in to github...")
	out, err := gh.exec("gh", "auth", "status")
	if err != nil {
		fmt.Printf(string(out))
		return false
	}

	log.Debug("User is logged in!")
	return true

}

func (gh GhCliClient) LogInToGh() error {
	log.Debug("Logging user in to github...")
	_, err := gh.exec("gh", "auth", "login")
	if err != nil {
		return err
	}

	return nil
}

func (gh GhCliClient) IsValidGhRepo(repo string) error {
	_, err := gh.exec("gh", "repo", "view", repo)
	if err != nil {
		log.Debug("Github repo " + repo + "not found")
		return err
	}
	return nil
}

func (gh GhCliClient) GetRepoNameWithOwner() (string, error) {
	repoNameWithOwner := ""
	out, err := gh.exec("gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err != nil {
		log.Fatal("getting github repo name with owner")
		return repoNameWithOwner, err
	}
	if out == "" {
		log.Fatal("github repo name empty from gh cli")
		return repoNameWithOwner, fmt.Errorf("github repo name empty from gh cli")
	}

	repoNameWithOwner = string(out)
	repoNameWithOwner = strings.TrimSpace(repoNameWithOwner)
	log.Debug("retrieved repoNameWithOwner from gh cli: ", repoNameWithOwner)
	return repoNameWithOwner, nil
}
