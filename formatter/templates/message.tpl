[[{{ .Message.Name }}_message]]
## {{ .Message.Name }} Message
{{ GetComments .Message.Comments }}

.{{ .Message.Name }} Overview
{{ if eq (index .Parameters "collapsible") "on" -}}
[%collapsible]
====
{{ end -}}
{{ $commentSize := 3 -}}
{{ if eq (index .Parameters "rest") "on" -}}
[cols="2,2,1,1,1", options="header"]
|===
| Name | Type | Repeated | Sequence | JSON Name
{{$commentSize = 4 -}}
{{else -}}
[cols="2,2,1,1", options="header"]
|===
| Name | Type | Repeated | Sequence
{{ end -}}
{{- /* we must declare the Message name and Parameters to a var because of scoping */ -}}
{{ $messageName := .Message.Name -}}
{{ $parameters := .Parameters -}}
{{ range .Message.Fields -}}
|[[{{ .Name}}_{{$messageName}}]] {{.Name}}
| {{ GetFieldType . $parameters }}
| {{ BoolIcon .Repeated $parameters }}
| {{ .Number }}
{{ if eq (index $parameters "rest") "on" -}}
| {{ .JsonName }}
{{ end -}}
{{ if .HasComments -}}
| 
{{ $commentSize }}+| {{ GetComments .Comments }}

{{ end -}}
{{ else -}}

{{ end -}}
|===

{{ if eq (index .Parameters "collapsible") "on" -}}
====

{{ end -}}