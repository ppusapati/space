{{/*
Common helpers for the P9E Space umbrella chart.
*/}}

{{/*
Image reference. Honors global.imageRegistry, falls back to per-service
image.repository. Tag is global.imageTag unless overridden per service.
*/}}
{{- define "space.image" -}}
{{- $svc := index . 0 -}}
{{- $globals := index . 1 -}}
{{- $registry := default "" $globals.imageRegistry -}}
{{- $repo := default (printf "p9e/%s" (index . 2)) $svc.image.repository -}}
{{- $tag := default $globals.imageTag (default "" $svc.image.tag) -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- else -}}
{{- printf "%s:%s" $repo $tag -}}
{{- end -}}
{{- end -}}

{{/*
Common labels applied to every Kubernetes object.
*/}}
{{- define "space.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/part-of: space
{{- end -}}

{{/*
Per-service selector labels.
*/}}
{{- define "space.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .release }}
{{- end -}}

{{/*
ServiceAccount name resolver.
*/}}
{{- define "space.serviceAccountName" -}}
{{- if .Values.global.serviceAccount.create -}}
{{- default (printf "%s-sa" .Release.Name) .Values.global.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.global.serviceAccount.name -}}
{{- end -}}
{{- end -}}
