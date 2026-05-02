{{/*
chetana.networkpolicy — default-deny ingress with explicit allows from
.Values.networkPolicy.ingress. The empty-array case (no allowed
sources) collapses to a NetworkPolicy that selects the pods but lists
no ingress rules; Kubernetes interprets that as "deny all ingress",
which is the safe default required by REQ-NFR-REL-004.
*/}}
{{- define "chetana.networkpolicy" -}}
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "chetana.fullname" . }}
  labels:
    {{- include "chetana.labels" . | nindent 4 }}
spec:
  podSelector:
    matchLabels:
      {{- include "chetana.selectorLabels" . | nindent 6 }}
  policyTypes:
    - Ingress
    {{- if .Values.networkPolicy.egress }}
    - Egress
    {{- end }}
  {{- if .Values.networkPolicy.ingress }}
  ingress:
    {{- toYaml .Values.networkPolicy.ingress | nindent 4 }}
  {{- else }}
  # Empty ingress = default-deny. REQ-NFR-REL-004.
  ingress: []
  {{- end }}
  {{- with .Values.networkPolicy.egress }}
  egress:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end -}}
