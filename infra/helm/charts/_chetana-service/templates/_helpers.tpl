{{/*
_helpers.tpl — chart-local helper functions.

Naming conventions follow the canonical Helm idiom: "chetana.fullname",
"chetana.labels", etc. Service charts include this library via:

    dependencies:
      - name: chetana-service
        repository: file://../../charts/_chetana-service
        version: 0.1.0

…and then `{{ include "chetana.deployment" . }}` etc. in their own
templates.
*/}}

{{/*
chetana.fullname — release-name + service-name, truncated to 63 chars
to satisfy Kubernetes DNS-1123 label constraints.
*/}}
{{- define "chetana.fullname" -}}
{{- $name := .Values.service.name | default .Chart.Name -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
chetana.labels — common labels applied to every resource.
The chetana_region label drives multi-region scheduling.
*/}}
{{- define "chetana.labels" -}}
app.kubernetes.io/name: {{ .Values.service.name | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion | default "unknown" | quote }}
chetana.p9e.in/region: {{ .Values.region | quote }}
{{- end -}}

{{/*
chetana.selectorLabels — narrow label selector used by Service +
Deployment + HPA + PDB. Stable across image bumps so rolling updates
keep the selector in place.
*/}}
{{- define "chetana.selectorLabels" -}}
app.kubernetes.io/name: {{ .Values.service.name | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
{{- end -}}

{{/*
chetana.serviceAccountName — derived once and reused. Defaults to the
fullname when serviceAccount.name is unset.
*/}}
{{- define "chetana.serviceAccountName" -}}
{{- if and .Values.serviceAccount (hasKey .Values.serviceAccount "name") (ne .Values.serviceAccount.name "") -}}
{{ .Values.serviceAccount.name }}
{{- else -}}
{{ include "chetana.fullname" . }}
{{- end -}}
{{- end -}}
