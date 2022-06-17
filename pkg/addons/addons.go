package addons

import (
	"embed"
	"errors"
	"fmt"
	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

//go:generate cp -r ../../addons ./addons

var (
	//go:embed addons
	addons        embed.FS
	parentDirName = "addons"
)

type AddOn struct {
	templates fs.DirEntry
	dest      string
}

func GenerateAddon(provider, addon, dest string, userInputs map[string]string) error {
	providerPath := parentDirName + "/" + strings.ToLower(provider)
	addonMap, err := embedutils.EmbedFStoMap(addons, providerPath)
	if err != nil {
		return nil
	}
	if addon == "" {
		addonNames := getKeySet(addonMap)
		prompt := promptui.Select{
			Label: fmt.Sprintf("Select %s addon", provider),
			Items: addonNames,
		}
		_, addon, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	selectedAddon := addonMap[addon]
	if selectedAddon == nil {
		return errors.New("addon not found")
	}

	addonVals, err := getAddonValues(providerPath+"/"+selectedAddon.Name(), dest, userInputs)
	if err != nil {
		return err
	}

	log.Debugf("addonVals: %s", addonVals)
	return err
}

func getAddonValues(selectedAddonPath, dest string, userInputs map[string]string) ([]map[string]string, error) {
	configBytes, err := fs.ReadFile(addons, selectedAddonPath+"/draft_config.yaml")
	if err != nil {
		return nil, err
	}

	var addOnConfig config.AddonConfig
	if err = yaml.Unmarshal(configBytes, &addOnConfig); err != nil {
		return nil, err
	}

	if userInputs == nil {
		userInputs, err = prompts.RunPromptsFromConfig(&addOnConfig.DraftConfig)
		if err != nil {
			return nil, err
		}
	}

	referenceMap, err := addOnConfig.GetReferenceMap(dest)
	if err != nil {
		return nil, err
	}

	addonVals := []map[string]string{userInputs, referenceMap}

	return addonVals, nil
}

func getKeySet[K comparable, V any](aMap map[K]V) []K {
	keys := make([]K, 0, len(aMap))
	for key := range aMap {
		keys = append(keys, key)
	}
	return keys
}
