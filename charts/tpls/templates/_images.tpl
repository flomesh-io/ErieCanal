{{/* pipy image without repository */}}
{{- define "ErieCanal.pipy.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ErieCanal.pipy.imageName .Values.ErieCanal.pipy.tag -}}
{{- end -}}

{{/* pipy image */}}
{{- define "ErieCanal.pipy.image" -}}
{{- printf "%s/%s" .Values.ErieCanal.image.repository (include "ErieCanal.pipy.image.wo-repo" .) -}}
{{- end -}}

{{/* pipy-repo image without repository */}}
{{- define "ErieCanal.pipy-repo.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ErieCanal.pipyRepo.imageName .Values.ErieCanal.pipyRepo.tag -}}
{{- end -}}

{{/* pipy-repo image */}}
{{- define "ErieCanal.pipy-repo.image" -}}
{{- printf "%s/%s" .Values.ErieCanal.image.repository (include "ErieCanal.pipy-repo.image.wo-repo" .) -}}
{{- end -}}

{{/* wait-for-it image without repository */}}
{{- define "ErieCanal.wait-for-it.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ErieCanal.waitForIt.imageName .Values.ErieCanal.waitForIt.tag -}}
{{- end -}}

{{/* wait-for-it image */}}
{{- define "ErieCanal.wait-for-it.image" -}}
{{- printf "%s/%s" .Values.ErieCanal.image.repository (include "ErieCanal.wait-for-it.image.wo-repo" .) -}}
{{- end -}}

{{/* toolbox image without repository */}}
{{- define "ErieCanal.toolbox.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ErieCanal.toolbox.imageName .Values.ErieCanal.toolbox.tag -}}
{{- end -}}

{{/* toolbox image */}}
{{- define "ErieCanal.toolbox.image" -}}
{{- printf "%s/%s" .Values.ErieCanal.image.repository (include "ErieCanal.toolbox.image.wo-repo" .) -}}
{{- end -}}

{{/* proxy-init image without repository */}}
{{- define "ErieCanal.proxy-init.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ErieCanal.proxyInit.name (include "ErieCanal.app-version" .) -}}
{{- end -}}

{{/* proxy-init image */}}
{{- define "ErieCanal.proxy-init.image" -}}
{{- printf "%s/%s" .Values.ErieCanal.image.repository (include "ErieCanal.proxy-init.image.wo-repo" .) -}}
{{- end -}}

{{/* manager image */}}
{{- define "ErieCanal.manager.image" -}}
{{- printf "%s/%s:%s" .Values.ErieCanal.image.repository .Values.ErieCanal.manager.name (include "ErieCanal.app-version" .) -}}
{{- end -}}

{{/* ingress-pipy image */}}
{{- define "ErieCanal.ingress-pipy.image" -}}
{{- printf "%s/%s:%s" .Values.ErieCanal.image.repository .Values.ErieCanal.ingress.name (include "ErieCanal.app-version" .) -}}
{{- end -}}

{{/* curl image without repository */}}
{{- define "ErieCanal.curl.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ErieCanal.curl.imageName .Values.ErieCanal.curl.tag -}}
{{- end -}}

{{/* curl image */}}
{{- define "ErieCanal.curl.image" -}}
{{- printf "%s/%s" .Values.ErieCanal.image.repository (include "ErieCanal.curl.image.wo-repo" .) -}}
{{- end -}}

{{/* service-lb image without repository */}}
{{- define "ErieCanal.service-lb.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ErieCanal.serviceLB.imageName .Values.ErieCanal.serviceLB.tag -}}
{{- end -}}

{{/* service-lb image */}}
{{- define "ErieCanal.service-lb.image" -}}
{{- printf "%s/%s" .Values.ErieCanal.image.repository (include "ErieCanal.service-lb.image.wo-repo" .) -}}
{{- end -}}