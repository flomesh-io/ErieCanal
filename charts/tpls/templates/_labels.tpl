{{/*
Common labels
*/}}
{{- define "ec.labels" -}}
helm.sh/chart: {{ include "ec.chart" . }}
app.kubernetes.io/version: {{ include "ec.app-version" . | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/name: {{ .Chart.Name }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ec.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ec.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Common labels - manager
*/}}
{{- define "ec.manager.labels" -}}
{{ include "ec.labels" . }}
app.kubernetes.io/component: erie-canal-manager
app.kubernetes.io/instance: erie-canal-manager
{{- end }}

{{/*
Selector labels - manager
*/}}
{{- define "ec.manager.selectorLabels" -}}
app: {{ .Values.ec.manager.name }}
flomesh.io/app: {{ .Values.ec.manager.name }}
{{- end }}

{{/*
Common labels - webhook-service
*/}}
{{- define "ec.webhook-service.labels" -}}
{{ include "ec.labels" . }}
app.kubernetes.io/component: erie-canal-webhook
app.kubernetes.io/instance: erie-canal-manager
{{- end }}

{{/*
Selector labels - webhook-service
*/}}
{{- define "ec.webhook-service.selectorLabels" -}}
{{ include "ec.manager.selectorLabels" . }}
{{- end }}

{{/*
Common labels - repo-service
*/}}
{{- define "ec.repo.labels" -}}
{{ include "ec.labels" . }}
app.kubernetes.io/component: erie-canal-repo
app.kubernetes.io/instance: erie-canal-repo
{{- end }}

{{/*
Selector labels - repo-service
*/}}
{{- define "ec.repo.selectorLabels" -}}
app: {{ .Values.ec.repo.name }}
flomesh.io/app: {{ .Values.ec.repo.name }}
{{- end }}

{{/*
Common labels - ingress-pipy
*/}}
{{- define "ec.ingress-pipy.labels" -}}
{{ include "ec.labels" . }}
app.kubernetes.io/component: controller
app.kubernetes.io/instance: erie-canal-ingress-pipy
{{- end }}

{{/*
Selector labels - ingress-pipy
*/}}
{{- define "ec.ingress-pipy.selectorLabels" -}}
app: {{ .Values.ec.ingress.name }}
flomesh.io/app: {{ .Values.ec.ingress.name }}
{{- end }}

{{/*
Common labels - egress-gateway
*/}}
{{- define "ec.egress-gateway.labels" -}}
{{ include "ec.labels" . }}
app.kubernetes.io/component: erie-canal-egress-gateway
app.kubernetes.io/instance: erie-canal-egress-gateway
{{- end }}

{{/*
Selector labels - egress-gateway
*/}}
{{- define "ec.egress-gateway.selectorLabels" -}}
app: {{ .Values.ec.egressGateway.name }}
flomesh.io/app: {{ .Values.ec.egressGateway.name }}
{{- end }}