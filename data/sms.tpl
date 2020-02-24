[{{ upper .Status }}{{ if eq .Status "firing" }}:{{ .Alerts | len }}{{ end }}] {{ .CommonLabels.alertname }}
Cause: {{ .CommonAnnotations.summary }}
Link: {{ .ExternalURL }}