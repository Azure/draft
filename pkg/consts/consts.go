package consts

var HelmReferencePathMapping = map[string]map[string]string{
	"service": {
		"metadata.name":   `{{ include "{{APPNAME}}.fullname" . }}`,
		"spec.ports.port": `{{ .Values.service.port }}`,
	},
}

var RefPathLookups = map[string]map[string][]string{
	"service": {
		"metadata.name":   []string{"metadata", "name"},
		"spec.ports.port": []string{"spec", "ports", "-", "port"},
	},
}
