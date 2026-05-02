{{/*
chetana.serviceaccount — ServiceAccount with optional IRSA annotations
(eks.amazonaws.com/role-arn) supplied by the consumer.
*/}}
{{- define "chetana.serviceaccount" -}}
{{- if and .Values.serviceAccount (or (not (hasKey .Values.serviceAccount "create")) .Values.serviceAccount.create) -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "chetana.serviceAccountName" . }}
  labels:
    {{- include "chetana.labels" . | nindent 4 }}
  {{- with .Values.serviceAccount.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end -}}
{{- end -}}
