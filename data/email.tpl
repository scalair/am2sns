Cause: {{ .CommonAnnotations.summary }}
Link: {{ .ExternalURL }}

Details:
{{ range $key, $value := .CommonLabels }}
- {{ $key }}: {{ $value }}
{{ end }}
