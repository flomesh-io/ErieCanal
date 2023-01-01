{{/*
Common labels - egress-gateway
*/}}
{{- define "ec.egress-gateway.annotations" -}}
openservicemesh.io/egress-gateway-mode: {{ .Values.ec.egressGateway.mode }}
{{- end }}