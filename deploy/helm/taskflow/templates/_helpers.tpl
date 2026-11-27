{{- define "taskflow.fullname" -}}
{{ .Release.Name }}
{{- end }}

{{- define "taskflow.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "taskflow.labels" -}}
{{ include "taskflow.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.image.tag | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end }}

{{- define "taskflow.image" -}}
{{ .Values.image.repository }}:{{ required "image.tag is required: pass --set image.tag=<commit sha>" .Values.image.tag }}
{{- end }}