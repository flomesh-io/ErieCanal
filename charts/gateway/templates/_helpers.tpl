{{/*
ServiceAccountName - namespaced-ingress
*/}}
{{- define "ec.gateway.serviceAccountName" -}}
{{ default "erie-canal-gateway" .Values.gateway.spec.serviceAccountName }}
{{- end }}


{{- define "ec.namespaced-ingress.heath.port" -}}
{{- if and .Values.ec.ingress.enabled .Values.ec.ingress.namespaced }}
{{- if .Values.nsig.spec.http.enabled }}
{{- default .Values.ec.ingress.http.containerPort .Values.nsig.spec.http.port.targetPort }}
{{- else if and .Values.nsig.spec.tls.enabled }}
{{- default .Values.ec.ingress.tls.containerPort .Values.nsig.spec.tls.port.targetPort }}
{{- else }}
8081
{{- end }}
{{- else }}
8081
{{- end }}
{{- end }}
