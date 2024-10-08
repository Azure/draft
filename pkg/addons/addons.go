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

func GenerateAddon(addons embed.FS, provider, addon, dest string, addonConfig AddonConfig, templateWriter templatewriter.TemplateWriter) error {
	selectedAddonPath, err := GetAddonPath(addons, provider, addon)
	if err != nil {
		return err
	}

	addonDestPath, err := addonConfig.GetAddonDestPath(dest)
	if err != nil {
		return err
	}

	if addonConfig.DraftConfig == nil {
		return errors.New("DraftConfig is nil")
	} else {
		err := addonConfig.DraftConfig.ApplyDefaultVariables()
		if err != nil {
			return err
		}
	}

	if err = osutil.CopyDirWithTemplates(addons, selectedAddonPath, addonDestPath, addonConfig.DraftConfig, templateWriter); err != nil {
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

	addonConfigPath := path.Join(selectedAddonPath, "draft.yaml")
	log.Debugf("addonConfig is: %s", addonConfigPath)

	configBytes, err := fs.ReadFile(addons, addonConfigPath)
	if err != nil {
		return AddonConfig{}, err
	}
	var addonConfig AddonConfig
	if err = yaml.Unmarshal(configBytes, &addonConfig); err != nil {
		return AddonConfig{}, err
	}

	return addonConfig, nil
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

func PromptAddonValues(dest string, addonConfig *AddonConfig) error {
	err := prompts.RunPromptsFromConfigWithSkips(addonConfig.DraftConfig)
	if err != nil {
		return err
	}
	log.Debug("got user inputs")

	referenceMap, err := addonConfig.GetReferenceValueMap(dest)
	if err != nil {
		return err
	}
	log.Debug("got reference map")
	// merge maps
	for refName, refVal := range referenceMap {
		// check for key collision
		if _, err := addonConfig.DraftConfig.GetVariable(refName); err == nil {
			return errors.New("variable name collision between references and DraftConfig")
		}
		if strings.Contains(strings.ToLower(refName), "namespace") && refVal == "" {
			refVal = "default" //hack here to have explicit namespacing, probably a better way to do this
		}
		addonConfig.DraftConfig.SetVariable(refName, refVal)
	}

	return nil
}
