package config

import (
	log "github.com/sirupsen/logrus"
)

type DraftConfig struct {
	NameOverrides    []FileNameOverride
	Variables        []BuilderVar
	nameOverrideMap  map[string]string
	VariableDefaults []BuilderVarDefault
}

type FileNameOverride struct {
	Path   string
	Prefix string
}

type BuilderVar struct {
	Name        string
	Description string
	VarType     string
}

type BuilderVarDefault struct {
	Name  string
	Value string
}

func (d *DraftConfig) initialize() {
	d.nameOverrideMap = make(map[string]string)
	log.Debug("initializing nameOverrideMap")
	for _, builderVar := range d.NameOverrides {
		log.Debugf("mapping path: %s, to prefix %s", builderVar.Path, builderVar.Prefix)
		d.nameOverrideMap[builderVar.Path] = builderVar.Prefix
	}
}

func (d *DraftConfig) GetNameOverride(path string) string {
	if d.nameOverrideMap == nil {
		d.initialize()
	}
	prefix, ok := d.nameOverrideMap[path]
	if !ok {
		return ""
	}

	return prefix
}
