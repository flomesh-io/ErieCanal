{{/*
Service Host - repo-service
*/}}
{{- define "ErieCanal.repo-service.host" -}}
{{- if .Values.ErieCanal.repo.preProvision.enabled }}
{{- .Values.ErieCanal.repo.preProvision.host }}
{{- else }}
{{- printf "%s.%s.svc" .Values.ErieCanal.services.repo.name (include "ErieCanal.namespace" .) -}}
{{- end }}
{{- end }}

{{/*
Service Port - repo-service
*/}}
{{- define "ErieCanal.repo-service.port" -}}
{{- if .Values.ErieCanal.repo.preProvision.enabled }}
{{- .Values.ErieCanal.repo.preProvision.port }}
{{- else }}
{{- .Values.ErieCanal.services.repo.port }}
{{- end }}
{{- end }}

{{/*
Service Address - repo-service
*/}}
{{- define "ErieCanal.repo-service.addr" -}}
{{- printf "%s:%s" (include "ErieCanal.repo-service.host" .) (include "ErieCanal.repo-service.port" .) -}}
{{- end }}

{{/*
Service URL(http) - repo-service
*/}}
{{- define "ErieCanal.repo-service.url" -}}
{{- printf "%s://%s" .Values.ErieCanal.repo.schema (include "ErieCanal.repo-service.addr" .) -}}
{{- end }}

{{/*
Service Host - webhook-service
*/}}
{{- define "ErieCanal.webhook-service.host" -}}
{{- printf "%s.%s.svc" .Values.ErieCanal.services.webhook.name (include "ErieCanal.namespace" .) -}}
{{- end }}

{{/*
Service Address - webhook-service
*/}}
{{- define "ErieCanal.webhook-service.addr" -}}
{{- printf "%s:%d" (include "ErieCanal.webhook-service.host" .) (int .Values.ErieCanal.services.webhook.port) -}}
{{- end }}

{{/*
Service Full Name - manager
*/}}
{{- define "ErieCanal.manager.host" -}}
{{- printf "%s.%s.svc" .Values.ErieCanal.services.manager.name (include "ErieCanal.namespace" .) -}}
{{- end }}