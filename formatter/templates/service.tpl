[[{{ .Service.Name }}_service]]
## {{ .Service.Name }} Service
{{ GetComments .Service.Comments }}

{{ range .Service.Methods }}
### {{ .Name }}
{{ GetComments .Comments }}

|===
| Request Type | {{ .InputType }}
| Return Type | {{ .OutputType }}
|====
{{ end }}