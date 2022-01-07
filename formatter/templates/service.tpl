[[{{ .Service.Name }}_service]]
## {{ .Service.Name }} Service
{{ GetComments .Service.Comments }}
{{ $parameters := .Parameters -}}
{{ range .Service.Methods }}
### {{ .Name }}
{{ GetComments .Comments }}

|===
| Request Type      | {{ GetTypeFromString .InputType $parameters }}
| Request Streaming | {{ BoolIcon .ClientStreaming $parameters }}
| Return Type       | {{ GetTypeFromString .OutputType $parameters }}
| Return Streaming  | {{ BoolIcon .ServerStreaming $parameters }}
| REST Support      | {{ BoolIcon .RESTSupport $parameters }}
{{ if .RESTSupport -}}
| REST Method       | {{ .RESTMethod }}
| REST URL          | {{ .RESTURL }}
{{ end -}}
|===


{{ end }}