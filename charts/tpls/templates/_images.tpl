{{/* pipy image without repository */}}
{{- define "ec.pipy.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ec.pipy.imageName .Values.ec.pipy.tag -}}
{{- end -}}

{{/* pipy image */}}
{{- define "ec.pipy.image" -}}
{{- printf "%s/%s" .Values.ec.image.repository (include "ec.pipy.image.wo-repo" .) -}}
{{- end -}}

{{/* pipy-repo image without repository */}}
{{- define "ec.pipy-repo.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ec.pipyRepo.imageName .Values.ec.pipyRepo.tag -}}
{{- end -}}

{{/* pipy-repo image */}}
{{- define "ec.pipy-repo.image" -}}
{{- printf "%s/%s" .Values.ec.image.repository (include "ec.pipy-repo.image.wo-repo" .) -}}
{{- end -}}

{{/* wait-for-it image without repository */}}
{{- define "ec.wait-for-it.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ec.waitForIt.imageName .Values.ec.waitForIt.tag -}}
{{- end -}}

{{/* wait-for-it image */}}
{{- define "ec.wait-for-it.image" -}}
{{- printf "%s/%s" .Values.ec.image.repository (include "ec.wait-for-it.image.wo-repo" .) -}}
{{- end -}}

{{/* toolbox image without repository */}}
{{- define "ec.toolbox.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ec.toolbox.imageName .Values.ec.toolbox.tag -}}
{{- end -}}

{{/* toolbox image */}}
{{- define "ec.toolbox.image" -}}
{{- printf "%s/%s" .Values.ec.image.repository (include "ec.toolbox.image.wo-repo" .) -}}
{{- end -}}

{{/* proxy-init image without repository */}}
{{- define "ec.proxy-init.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ec.proxyInit.name (include "ec.app-version" .) -}}
{{- end -}}

{{/* proxy-init image */}}
{{- define "ec.proxy-init.image" -}}
{{- printf "%s/%s" .Values.ec.image.repository (include "ec.proxy-init.image.wo-repo" .) -}}
{{- end -}}

{{/* manager image */}}
{{- define "ec.manager.image" -}}
{{- printf "%s/%s:%s" .Values.ec.image.repository .Values.ec.manager.name (include "ec.app-version" .) -}}
{{- end -}}

{{/* ingress-pipy image */}}
{{- define "ec.ingress-pipy.image" -}}
{{- printf "%s/%s:%s" .Values.ec.image.repository .Values.ec.ingress.name (include "ec.app-version" .) -}}
{{- end -}}

{{/* curl image without repository */}}
{{- define "ec.curl.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ec.curl.imageName .Values.ec.curl.tag -}}
{{- end -}}

{{/* curl image */}}
{{- define "ec.curl.image" -}}
{{- printf "%s/%s" .Values.ec.image.repository (include "ec.curl.image.wo-repo" .) -}}
{{- end -}}

{{/* service-lb image without repository */}}
{{- define "ec.service-lb.image.wo-repo" -}}
{{- printf "%s:%s" .Values.ec.serviceLB.imageName .Values.ec.serviceLB.tag -}}
{{- end -}}

{{/* service-lb image */}}
{{- define "ec.service-lb.image" -}}
{{- printf "%s/%s" .Values.ec.image.repository (include "ec.service-lb.image.wo-repo" .) -}}
{{- end -}}