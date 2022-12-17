{{/* common envs */}}
{{- define "ec.common-env" -}}
{{- with .Values.ec.commonEnv }}
{{- toYaml . }}
{{- end }}
- name: ERIECANAL_NAMESPACE
  value: {{ include "ec.namespace" . }}
{{- end -}}