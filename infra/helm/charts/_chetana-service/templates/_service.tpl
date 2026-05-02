{{/*
chetana.service — exposes both the RPC port and the metrics port. The
metrics port carries a separate name so prometheus-operator can target
it via a ServiceMonitor selector matching `port: metrics`.
*/}}
{{- define "chetana.service" -}}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "chetana.fullname" . }}
  labels:
    {{- include "chetana.labels" . | nindent 4 }}
spec:
  type: ClusterIP
  selector:
    {{- include "chetana.selectorLabels" . | nindent 4 }}
  ports:
    - name: rpc
      port: {{ .Values.service.port }}
      targetPort: rpc
      protocol: TCP
    - name: metrics
      port: {{ .Values.service.metricsPort }}
      targetPort: metrics
      protocol: TCP
{{- end -}}
