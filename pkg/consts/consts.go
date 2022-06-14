package consts

var HelmReferencePathMapping = map[string]map[string]string{
	"service": {
		"metadata.name":   `{{ include "{{APPNAME}}.fullname" . }}`,
		"spec.ports.port": `{{ .Values.service.port }}`,
	},
}
