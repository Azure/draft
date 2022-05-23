package config

import "github.com/Azure/draft/pkg/types"

type AddonConfig struct {
	DraftConfig
	References map[string]reference
}

type reference struct {
	Name string
	Path string
}

func (ac *AddonConfig) GetReferenceMap(referenceType, referencePath string) (map[string]string, error) {

}

func (ac *AddonConfig) getServiceReference(svc types.ServiceManifest) map[string]string {
	
}
