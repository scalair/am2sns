Cause: {{ .CommonAnnotations.summary }}
Link: {{ .ExternalURL }}

Details:
{{ range $key, $value := .CommonLabels }}
- {{ $key }}: {{ $value }}
{{ end }}

Alerts:
{{ range $alert := .Alerts }}
-  {{ $alert.Annotations.message }}
{{ end }}
