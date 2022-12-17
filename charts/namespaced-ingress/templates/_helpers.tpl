{{/*
ServiceAccountName - namespaced-ingress
*/}}
{{- define "ec.namespaced-ingress.serviceAccountName" -}}
{{ default "erie-canal-namespaced-ingress" .Values.nsig.spec.serviceAccountName }}
{{- end }}
