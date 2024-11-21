{{/*
Expand the name of the chart.
*/}}
{{ .Config.GetVariableValue "APPNAME" | printf "{{- define \"%s.name\" -}}"  }}
{{`{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}`}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{ .Config.GetVariableValue "APPNAME" | printf "{{- define \"%s.fullname\" -}}" }}
{{`{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}`}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{ .Config.GetVariableValue "APPNAME" | printf "{{- define \"%s.chart\" -}}" }}
{{`{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}`}}

{{/*
Common labels
*/}}
{{ .Config.GetVariableValue "APPNAME" | printf "{{- define \"%s.labels\" -}}" }}
helm.sh/chart: {{ .Config.GetVariableValue "APPNAME" | printf "{{ include \"%s.chart\" . }}" }}
{{ .Config.GetVariableValue "APPNAME" | printf "{{ include \"%s.selectorLabels\" . }}" }}
kubernetes.azure.com/generator: {{ printf "{{ .Values.generatorLabel }}" }}
{{`{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}`}}

{{/*
Selector labels
*/}}
{{ .Config.GetVariableValue "APPNAME" | printf "{{- define \"%s.selectorLabels\" -}}" }}
{{ .Config.GetVariableValue "APPNAME" | printf "app.kubernetes.io/name: {{ include \"%s.name\" . }}\napp.kubernetes.io/instance: {{ .Release.Name }}\n{{- end }}" }}