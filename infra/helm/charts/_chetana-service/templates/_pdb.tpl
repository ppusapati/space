{{/*
chetana.pdb — PodDisruptionBudget. The schema requires minAvailable to
be set, satisfying REQ-NFR-REL-003 (every service has a PDB).
*/}}
{{- define "chetana.pdb" -}}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "chetana.fullname" . }}
  labels:
    {{- include "chetana.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "chetana.selectorLabels" . | nindent 6 }}
  {{- if hasKey .Values.pdb "maxUnavailable" }}
  maxUnavailable: {{ .Values.pdb.maxUnavailable }}
  {{- else }}
  minAvailable: {{ .Values.pdb.minAvailable }}
  {{- end }}
{{- end -}}
