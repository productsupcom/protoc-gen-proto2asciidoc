[[{{ .Enum.Name }}_enum]]
## {{ .Enum.Name }} Enum
{{ GetComments .Enum.Comments }}

.{{ .Enum.Name }} Overview
{{ if eq (index .Parameters "collapsible") "on" -}}
[%collapsible]
====
{{ end -}}
{{ $commentSize := 3 -}}
[cols="2,1,3", options="header"]
|===
| Name | Sequence | Comment
{{- /* we must declare the Message name and Parameters to a var because of scoping */ -}}
{{ $enumName := .Enum.Name }}
{{ range .Enum.Values -}}
|[[{{ .Name}}_{{$enumName}}]] {{.Name}}
| {{ .Number }}
| {{ GetComments .Comments }}
{{ end -}}
|===

{{ if eq (index .Parameters "collapsible") "on" -}}
====

{{ end }}