{{/* common envs */}}
{{- define "ErieCanal.common-env" -}}
{{- with .Values.ErieCanal.commonEnv }}
{{- toYaml . }}
{{- end }}
- name: ERIECANAL_NAMESPACE
  value: {{ include "ErieCanal.namespace" . }}
{{- end -}}