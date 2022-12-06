{{/*
Common labels
*/}}
{{- define "ErieCanal.labels" -}}
helm.sh/chart: {{ include "ErieCanal.chart" . }}
app.kubernetes.io/version: {{ include "ErieCanal.app-version" . | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/name: {{ .Chart.Name }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ErieCanal.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ErieCanal.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common labels - manager
*/}}
{{- define "ErieCanal.manager.labels" -}}
{{ include "ErieCanal.labels" . }}
app.kubernetes.io/component: erie-canal-manager
app.kubernetes.io/instance: erie-canal-manager
{{- end }}

{{/*
Selector labels - manager
*/}}
{{- define "ErieCanal.manager.selectorLabels" -}}
app: {{ .Values.ErieCanal.manager.name }}
flomesh.io/app: {{ .Values.ErieCanal.manager.name }}
{{- end }}

{{/*
Common labels - webhook-service
*/}}
{{- define "ErieCanal.webhook-service.labels" -}}
{{ include "ErieCanal.labels" . }}
app.kubernetes.io/component: erie-canal-webhook
app.kubernetes.io/instance: erie-canal-manager
{{- end }}

{{/*
Selector labels - webhook-service
*/}}
{{- define "ErieCanal.webhook-service.selectorLabels" -}}
{{ include "ErieCanal.manager.selectorLabels" . }}
{{- end }}

{{/*
Common labels - repo-service
*/}}
{{- define "ErieCanal.repo.labels" -}}
{{ include "ErieCanal.labels" . }}
app.kubernetes.io/component: erie-canal-repo
app.kubernetes.io/instance: erie-canal-repo
{{- end }}

{{/*
Selector labels - repo-service
*/}}
{{- define "ErieCanal.repo.selectorLabels" -}}
app: {{ .Values.ErieCanal.repo.name }}
flomesh.io/app: {{ .Values.ErieCanal.repo.name }}
{{- end }}

{{/*
Common labels - ingress-pipy
*/}}
{{- define "ErieCanal.ingress-pipy.labels" -}}
{{ include "ErieCanal.labels" . }}
app.kubernetes.io/component: controller
app.kubernetes.io/instance: erie-canal-ingress-pipy
{{- end }}

{{/*
Selector labels - ingress-pipy
*/}}
{{- define "ErieCanal.ingress-pipy.selectorLabels" -}}
app: {{ .Values.ErieCanal.ingress.name }}
flomesh.io/app: {{ .Values.ErieCanal.ingress.name }}
{{- end }}

{{/*
Common labels - egress-gateway
*/}}
{{- define "ErieCanal.egress-gateway.labels" -}}
{{ include "ErieCanal.labels" . }}
app.kubernetes.io/component: erie-canal-egress-gateway
app.kubernetes.io/instance: erie-canal-egress-gateway
{{- end }}

{{/*
Selector labels - egress-gateway
*/}}
{{- define "ErieCanal.egress-gateway.selectorLabels" -}}
app: {{ .Values.ErieCanal.egressGateway.name }}
flomesh.io/app: {{ .Values.ErieCanal.egressGateway.name }}
{{- end }}