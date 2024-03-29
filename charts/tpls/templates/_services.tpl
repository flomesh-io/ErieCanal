{{/*
Service Host - repo-service
*/}}
{{- define "ec.repo-service.host" -}}
{{- if .Values.ec.repo.preProvision.enabled }}
{{- .Values.ec.repo.preProvision.host }}
{{- else }}
{{- printf "%s.%s.svc" .Values.ec.services.repo.name (include "ec.namespace" .) -}}
{{- end }}
{{- end }}

{{/*
Service Port - repo-service
*/}}
{{- define "ec.repo-service.port" -}}
{{- if .Values.ec.repo.preProvision.enabled }}
{{- .Values.ec.repo.preProvision.port }}
{{- else }}
{{- .Values.ec.services.repo.port }}
{{- end }}
{{- end }}

{{/*
Service Address - repo-service
*/}}
{{- define "ec.repo-service.addr" -}}
{{- printf "%s:%s" (include "ec.repo-service.host" .) (include "ec.repo-service.port" .) -}}
{{- end }}

{{/*
Service URL(http) - repo-service
*/}}
{{- define "ec.repo-service.url" -}}
{{- printf "%s://%s" .Values.ec.repo.schema (include "ec.repo-service.addr" .) -}}
{{- end }}

{{/*
Service Host - webhook-service
*/}}
{{- define "ec.webhook-service.host" -}}
{{- printf "%s.%s.svc" .Values.ec.services.webhook.name (include "ec.namespace" .) -}}
{{- end }}

{{/*
Service Address - webhook-service
*/}}
{{- define "ec.webhook-service.addr" -}}
{{- printf "%s:%d" (include "ec.webhook-service.host" .) (int .Values.ec.services.webhook.port) -}}
{{- end }}

{{/*
Service Full Name - manager
*/}}
{{- define "ec.manager.host" -}}
{{- printf "%s.%s.svc" .Values.ec.services.manager.name (include "ec.namespace" .) -}}
{{- end }}

{{/*
Service Host - ingress-pipy
*/}}
{{- define "ec.ingress-pipy.host" -}}
{{- printf "%s.%s.svc" .Values.ec.ingress.service.name (include "ec.namespace" .) -}}
{{- end }}

{{/*
Heath port - ingress-pipy
*/}}
{{- define "ec.ingress-pipy.heath.port" -}}
{{- if .Values.ec.ingress.enabled }}
{{- if and .Values.ec.ingress.http.enabled (not (empty .Values.ec.ingress.http.containerPort)) }}
{{- .Values.ec.ingress.http.containerPort }}
{{- else if and .Values.ec.ingress.tls.enabled (not (empty .Values.ec.ingress.tls.containerPort)) }}
{{- .Values.ec.ingress.tls.containerPort }}
{{- else }}
8081
{{- end }}
{{- else }}
8081
{{- end }}
{{- end }}