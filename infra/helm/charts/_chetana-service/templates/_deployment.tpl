{{/*
chetana.deployment — produces a Deployment with:
  • The standard chetana labels + region affinity hint.
  • Container ports for service.port (RPC) and service.metricsPort.
  • CHETANA_REGION env var injected from .Values.region so
    services/packages/region resolves correctly.
  • Liveness probe on /health, readiness probe on /ready.
*/}}
{{- define "chetana.deployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "chetana.fullname" . }}
  labels:
    {{- include "chetana.labels" . | nindent 4 }}
spec:
  {{/* When HPA is enabled the spec.replicas field is omitted so the
       HPA owns the scale. */}}
  {{- if not .Values.hpa.enabled }}
  replicas: {{ .Values.replicaCount | default 2 }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "chetana.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "chetana.selectorLabels" . | nindent 8 }}
        chetana.p9e.in/region: {{ .Values.region | quote }}
        {{- with .Values.podLabels }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: {{ .Values.service.metricsPort | quote }}
        prometheus.io/path: "/metrics"
        {{- with .Values.podAnnotations }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
    spec:
      serviceAccountName: {{ include "chetana.serviceAccountName" . }}
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: topology.kubernetes.io/region
                    operator: In
                    values: [{{ .Values.region | quote }}]
      containers:
        - name: {{ .Values.service.name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy | default "IfNotPresent" }}
          ports:
            - name: rpc
              containerPort: {{ .Values.service.port }}
              protocol: TCP
            - name: metrics
              containerPort: {{ .Values.service.metricsPort }}
              protocol: TCP
          env:
            - name: CHETANA_REGION
              value: {{ .Values.region | quote }}
            {{- with .Values.env }}
            {{- toYaml . | nindent 12 }}
            {{- end }}
          livenessProbe:
            httpGet:
              path: /health
              port: rpc
            initialDelaySeconds: 5
            periodSeconds: 10
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /ready
              port: rpc
            initialDelaySeconds: 3
            periodSeconds: 5
            failureThreshold: 3
          {{- with .Values.resources }}
          resources:
            {{- toYaml . | nindent 12 }}
          {{- end }}
{{- end -}}
