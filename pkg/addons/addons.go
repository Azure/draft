package addons

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/Azure/draft/pkg/templatewriter"
)

var (
	parentDirName = "addons"
)

func GenerateAddon(addons embed.FS, provider, addon, dest string, userInputs map[string]string, templateWriter templatewriter.TemplateWriter) error {
	providerPath := filepath.Join(parentDirName, strings.ToLower(provider))
	addonMap, err := embedutils.EmbedFStoMap(addons, providerPath)
	if err != nil {
		return err
	}
	if addon == "" {
		addonNames := maps.Keys(addonMap)
		prompt := promptui.Select{
			Label: fmt.Sprintf("Select %s addon", provider),
			Items: addonNames,
		}
		_, addon, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	var selectedAddon fs.DirEntry
	var ok bool
	if selectedAddon, ok = addonMap[addon]; !ok {
		return errors.New("addon not found")
	}

	selectedAddonPath := filepath.Join(providerPath, selectedAddon.Name())

	configBytes, err := fs.ReadFile(addons, filepath.Join(selectedAddonPath, "draft.yaml"))
	if err != nil {
		return err
	}

	var addOnConfig config.AddonConfig
	if err = yaml.Unmarshal(configBytes, &addOnConfig); err != nil {
		return err
	}

	log.Debugf("addOnConfig is: %s", addOnConfig)

	addonVals, err := getAddonValues(dest, userInputs, addOnConfig)
	if err != nil {
		return err
	}

	log.Debugf("addonValues are: %s", addonVals)

	addonDestPath, err := addOnConfig.GetAddonDestPath(dest)
	if err != nil {
		return err
	}

	if err = osutil.CopyDir(addons, selectedAddonPath, addonDestPath, &addOnConfig.DraftConfig, addonVals, templateWriter); err != nil {
		return err
	}

	return err
}

func getAddonValues(dest string, userInputs map[string]string, addOnConfig config.AddonConfig) (map[string]string, error) {
	log.Debugf("getAddonValues: %s", userInputs)
	var err error
	if userInputs == nil {
		userInputs = make(map[string]string)
	}

	if len(userInputs) == 0 {
		userInputs, err = prompts.RunPromptsFromConfig(&addOnConfig.DraftConfig)
		if err != nil {
			return nil, err
		}
		log.Debug("got user inputs")
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
