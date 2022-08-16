package config

import (
	log "github.com/sirupsen/logrus"
)

//TODO: remove Name Overrides since we don't need them anymore
type DraftConfig struct {
	NameOverrides    []FileNameOverride  `yaml:"nameOverrides"`
	Variables        []BuilderVar        `yaml:"variables"`
	VariableDefaults []BuilderVarDefault `yaml:"variableDefaults"`

	nameOverrideMap map[string]string
}

type FileNameOverride struct {
	Path   string `yaml:"path"`
	Prefix string `yaml:"prefix"`
}

type BuilderVar struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	VarType     string `yaml:"type"`
}

type BuilderVarDefault struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func (d *DraftConfig) initNameOverrideMap() {
	d.nameOverrideMap = make(map[string]string)
	log.Debug("initializing nameOverrideMap")
	for _, builderVar := range d.NameOverrides {
		log.Debugf("mapping path: %s, to prefix %s", builderVar.Path, builderVar.Prefix)
		d.nameOverrideMap[builderVar.Path] = builderVar.Prefix
	}
}

func (d *DraftConfig) GetNameOverride(path string) string {
	if d.nameOverrideMap == nil {
		d.initNameOverrideMap()
	}
	prefix, ok := d.nameOverrideMap[path]
	if !ok {
		return ""
	}

	return prefix
}
