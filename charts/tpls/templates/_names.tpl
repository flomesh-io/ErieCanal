{{/* Determine ErieCanal namespace */}}
{{- define "ec.namespace" -}}
{{- default .Release.Namespace .Values.ec.namespace }}
{{- end -}}

{{/*
Expand the name of the chart.
*/}}
{{- define "ec.name" -}}
{{- default .Chart.Name .Values.ec.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ec.fullname" -}}
{{- if .Values.ec.fullnameOverride }}
{{- .Values.ec.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.ec.nameOverride }}
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
{{- define "ec.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ec.serviceAccountName" -}}
{{- if .Values.ec.serviceAccount.create }}
{{- default .Chart.Name .Values.ec.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.ec.serviceAccount.name }}
{{- end }}
{{- end }}

{{/* Determine ErieCanal version */}}
{{- define "ec.app-version" -}}
{{- default .Chart.AppVersion .Values.ec.version }}
{{- end -}}