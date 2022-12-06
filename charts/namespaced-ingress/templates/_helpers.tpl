{{/*
ServiceAccountName - namespaced-ingress
*/}}
{{- define "ErieCanal.namespaced-ingress.serviceAccountName" -}}
{{ default "erie-canal-namespaced-ingress" .Values.nsig.spec.serviceAccountName }}
{{- end }}
