{{/* common envs */}}
{{- define "ErieCanal.common-env" -}}
{{- with .Values.ErieCanal.commonEnv }}
{{- toYaml . }}
{{- end }}
- name: ERIE_CANAL_NAMESPACE
  value: {{ include "ErieCanal.namespace" . }}
{{- end -}}