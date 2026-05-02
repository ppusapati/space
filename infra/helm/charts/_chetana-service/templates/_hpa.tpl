{{/*
chetana.hpa — HorizontalPodAutoscaler. Emitted only when
hpa.enabled=true; consumers MUST set this explicitly per the schema.
REQ-NFR-REL-003.
*/}}
{{- define "chetana.hpa" -}}
{{- if .Values.hpa.enabled -}}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "chetana.fullname" . }}
  labels:
    {{- include "chetana.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "chetana.fullname" . }}
  minReplicas: {{ .Values.hpa.minReplicas }}
  maxReplicas: {{ .Values.hpa.maxReplicas }}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.hpa.targetCPUUtilizationPercentage | default 70 }}
    {{- if .Values.hpa.targetMemoryUtilizationPercentage }}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{ .Values.hpa.targetMemoryUtilizationPercentage }}
    {{- end }}
{{- end -}}
{{- end -}}
