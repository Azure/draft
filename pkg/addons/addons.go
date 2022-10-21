package addons

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/Azure/draft/pkg/templatewriter"
)

var (
	parentDirName = "addons"
)

func GenerateAddon(addons embed.FS, provider, addon, dest string, userInputs map[string]string, templateWriter templatewriter.TemplateWriter) error {
	addOnConfig, err := GetAddonConfig(addons, provider, addon)
	if err != nil {
		return err
	}
	log.Debugf("addOnConfig is: %s", addOnConfig)

	selectedAddonPath, err := GetAddonPath(addons, provider, addon)
	if err != nil {
		return err
	}

	addonDestPath, err := addOnConfig.GetAddonDestPath(dest)
	if err != nil {
		return err
	}

	if err = osutil.CopyDir(addons, selectedAddonPath, addonDestPath, &addOnConfig.DraftConfig, userInputs, templateWriter); err != nil {
		return err
	}

	return err
}

func GetAddonPath(addons embed.FS, provider, addon string) (string, error) {
	providerPath := path.Join(parentDirName, strings.ToLower(provider))
	addonMap, err := embedutils.EmbedFStoMap(addons, providerPath)
	if err != nil {
		return "", err
	}
	var selectedAddon fs.DirEntry
	var ok bool
	if selectedAddon, ok = addonMap[addon]; !ok {
		return "", errors.New("addon not found")
	}

	selectedAddonPath := path.Join(providerPath, selectedAddon.Name())

	return selectedAddonPath, nil
}

func GetAddonConfig(addons embed.FS, provider, addon string) (AddonConfig, error) {
	selectedAddonPath, err := GetAddonPath(addons, provider, addon)
	if err != nil {
		return AddonConfig{}, err
	}

	configBytes, err := fs.ReadFile(addons, path.Join(selectedAddonPath, "draft.yaml"))
	if err != nil {
		return AddonConfig{}, err
	}
	var addOnConfig AddonConfig
	if err = yaml.Unmarshal(configBytes, &addOnConfig); err != nil {
		return AddonConfig{}, err
	}

	return addOnConfig, nil
}

func PromptAddon(addons embed.FS, provider string) (string, error) {
	providerPath := path.Join(parentDirName, strings.ToLower(provider))
	addonMap, err := embedutils.EmbedFStoMap(addons, providerPath)
	if err != nil {
		return "", err
	}

	addonNames := maps.Keys(addonMap)
	prompt := promptui.Select{
		Label: fmt.Sprintf("Select %s addon", provider),
		Items: addonNames,
	}
	_, addon, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return addon, nil
}

func PromptAddonValues(dest string, userInputs map[string]string, addOnConfig AddonConfig) (map[string]string, error) {
	log.Debugf("getAddonValues: %s", userInputs)
	var err error

	inputsToSkip := maps.Keys(userInputs)
	log.Debugf("inputsToSkip: %s", inputsToSkip)
	promptInputs, err := prompts.RunPromptsFromConfigWithSkips(&addOnConfig.DraftConfig, inputsToSkip)
	if err != nil {
		return nil, err
	}
	log.Debug("got user inputs")
	for k, v := range promptInputs {
		userInputs[k] = v
	}

	referenceMap, err := addOnConfig.GetReferenceValueMap(dest)
	if err != nil {
		return nil, err
	}
	log.Debug("got reference map")
	// merge maps
	for refName, refVal := range referenceMap {
		// check for key collision
		if _, ok := userInputs[refName]; ok {
			return nil, errors.New("variable name collision between references and userInputs")
		}
		if strings.Contains(strings.ToLower(refName), "namespace") && refVal == "" {
			refVal = "default" //hack here to have explicit namespacing, probably a better way to do this
		}
		userInputs[refName] = refVal
	}

	log.Debugf("merged maps into: %s", userInputs)
	return userInputs, nil
}
