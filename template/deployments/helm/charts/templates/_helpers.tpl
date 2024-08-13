{{/*
Expand the name of the chart.
*/}}
{{ printf "{{- define \"%s.name\" -}}" .APPNAME }}
{{ printf "{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix \"-\" }}\n{{- end }}" }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{ printf "{{- define \"%s.fullname\" -}}" .APPNAME }}
{{ printf "{{- if .Values.fullnameOverride }}\n{{- .Values.fullnameOverride | trunc 63 | trimSuffix \"-\" }}\n{{- else }}\n{{- $name := default .Chart.Name .Values.nameOverride }}\n{{- if contains $name .Release.Name }}\n{{- .Release.Name | trunc 63 | trimSuffix \"-\" }}\n{{- else }}\n{{- printf \"%s-%s\" .Release.Name $name | trunc 63 | trimSuffix \"-\" }}\n{{- end }}\n{{- end }}\n{{- end }}" }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{ printf "{{- define \"%s.chart\" -}}" .APPNAME }}
{{ printf "{{- printf \"%s-%s\" .Chart.Name .Chart.Version | replace \"+\" \"_\" | trunc 63 | trimSuffix \"-\" }}\n{{- end }}" }}

{{/*
Common labels
*/}}
{{ printf "{{- define \"%s.labels\" -}}" .APPNAME }}
{{ printf "helm.sh/chart: {{ include \"%s.chart\" . }}\n{{ include \"%s.selectorLabels\" . }}\n{{- if .Chart.AppVersion }}\napp.kubernetes.io/version: {{ .Chart.AppVersion | quote }}\n{{- end }}\napp.kubernetes.io/managed-by: {{ .Release.Service }}\n{{- end }}" .APPNAME .APPNAME }}

{{/*
Selector labels
*/}}
{{ printf "{{- define \"%s.selectorLabels\" -}}" .APPNAME }}
{{ printf "app.kubernetes.io/name: {{ include \"%s.name\" . }}\napp.kubernetes.io/instance: {{ .Release.Name }}\n{{- end }}" .APPNAME }}