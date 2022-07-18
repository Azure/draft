package consts

var HelmReferencePathMapping = map[string]map[string]string{
	"service": {
		"metadata.name":      `{{ include "{{APPNAME}}.fullname" . }}`,
		"spec.ports.port":    `{{ .Values.service.port }}`,
		"metadata.namespace": "default",
	},
}

var RefPathLookups = map[string]map[string][]string{
	"service": {
		"metadata.name":      []string{"metadata", "name"},
		"spec.ports.port":    []string{"spec", "ports", "-", "port"},
		"metadata.namespace": []string{"metadata", "namespace"},
	},
}

var DeploymentFilePaths = map[string]string{
	"helm":      "charts/templates",
	"kustomize": "overlays/production",
	"manifests": "manifests",
}
