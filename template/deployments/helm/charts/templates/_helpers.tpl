{{/*
Expand the name of the chart.
*/}}
{{ printf "{{- define \"%s.name\" -}}" .Config.GetVariableValue ".APPNAME" }}
{{`{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}`}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{ printf "{{- define \"%s.fullname\" -}}" .Config.GetVariableValue "APPNAME" }}
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
{{ printf "{{- define \"%s.chart\" -}}" .Config.GetVariableValue "APPNAME" }}
{{`{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}`}}

{{/*
Common labels
*/}}
{{ printf "{{- define \"%s.labels\" -}}" .Config.GetVariableValue "APPNAME" }}
helm.sh/chart: {{ printf "{{ include \"%s.chart\" . }}" .Config.GetVariableValue "APPNAME" }}
{{ printf "{{ include \"%s.selectorLabels\" . }}" .Config.GetVariableValue "APPNAME" }}
{{`{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}`}}

{{/*
Selector labels
*/}}
{{ printf "{{- define \"%s.selectorLabels\" -}}" .Config.GetVariableValue "APPNAME" }}
{{ printf "app.kubernetes.io/name: {{ include \"%s.name\" . }}\napp.kubernetes.io/instance: {{ .Release.Name }}\n{{- end }}" .Config.GetVariableValue "APPNAME" }}