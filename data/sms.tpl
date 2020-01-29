[{{ .Status | toUpper }}{{ if eq .Status "firing" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .CommonLabels.alertname }}
Cause: {{ .CommonAnnotations.summary }}
Link: {{ .ExternalURL }}