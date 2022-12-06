{{/* Determine ErieCanal namespace */}}
{{- define "ErieCanal.namespace" -}}
{{- default .Release.Namespace .Values.ErieCanal.namespace }}
{{- end -}}

{{/*
Expand the name of the chart.
*/}}
{{- define "ErieCanal.name" -}}
{{- default .Chart.Name .Values.ErieCanal.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ErieCanal.fullname" -}}
{{- if .Values.ErieCanal.fullnameOverride }}
{{- .Values.ErieCanal.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.ErieCanal.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ErieCanal.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ErieCanal.serviceAccountName" -}}
{{- if .Values.ErieCanal.serviceAccount.create }}
{{- default .Chart.Name .Values.ErieCanal.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.ErieCanal.serviceAccount.name }}
{{- end }}
{{- end }}

{{/* Determine ErieCanal version */}}
{{- define "ErieCanal.app-version" -}}
{{- default .Chart.AppVersion .Values.ErieCanal.version }}
{{- end -}}