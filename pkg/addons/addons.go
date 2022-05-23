package addons

import (
	"embed"
	"fmt"
	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
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

func GenerateAddon(provider, addon, dest string) error {
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
	log.Infof("loading addon %s", selectedAddon)

	configBytes, err := ioutil.ReadFile(providerPath + "/draft_config.yaml")
	if err != nil {
		return err
	}

	var addOnConfig config.AddonConfig
	if err = yaml.Unmarshal(configBytes, addOnConfig); err != nil {
		return err
	}
	userInputs, err := prompts.RunPromptsFromConfig(&addOnConfig.DraftConfig)
	if err != nil {
		return err
	}

}

func getKeySet[K, V](aMap map[K]V) []K {
	keys := make([]K, 0, len(aMap))
	for key := range aMap {
		keys = append(keys, key)
	}
	return keys
}
