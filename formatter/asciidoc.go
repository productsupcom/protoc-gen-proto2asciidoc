package formatter

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/productsupcom/proto2asciidoc/kit"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	_ "embed"
)

type Asciidoc struct {
	Formatter
}

func (ad *Asciidoc) GetFiles(desc kit.Desc) []*pluginpb.CodeGeneratorResponse_File {
	var out strings.Builder

	for _, service := range desc.Services {
		out.WriteString(GetTableForService(service, desc.Parameters))
	}

	for _, msg := range desc.Messages {
		out.WriteString(GetTableForMessage(msg, desc.Parameters))
	}

	for _, enum := range desc.Enums {
		out.WriteString(GetTableForEnum(enum, desc.Parameters))
	}

	name := "docs.adoc"
	content := out.String()
	var files []*pluginpb.CodeGeneratorResponse_File
	files = append(files, &pluginpb.CodeGeneratorResponse_File{
		Name:    &name,
		Content: &content,
	})

	return files
}

func BoolIcon(val bool, p kit.Parameters) string {
	if param, ok := p["icons"]; ok && param == "on" {
		if val {
			return "{true-icon}"
		}
		return "{false-icon}"
	}
	if val {
		return "true"
	}
	return "false"
}

func GetFieldType(f kit.Field, p kit.Parameters) string {
	if f.TypeName != nil && f.Type != nil {
		fqtn := strings.Split(*f.TypeName, ".")
		if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
			if val, ok := p["extension"]; ok && val == "on" {
				return fmt.Sprintf("proto2asciidoc:message[%s]", fqtn[len(fqtn)-1])
			}
			return fmt.Sprintf("<<%s_message>>", fqtn[len(fqtn)-1])
		}
		if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
			if val, ok := p["extension"]; ok && val == "on" {
				return fmt.Sprintf("proto2asciidoc:enum[%s]", fqtn[len(fqtn)-1])
			}
			return fmt.Sprintf("<<%s_enum>>", fqtn[len(fqtn)-1])
		}
	}
	return strings.ToLower(strings.TrimLeft(f.Type.String(), "TYPE_"))
}

// it would be nicer to ship this with the template as a file for in-promptu updates
// however that would require the user side to figure out paths etc
//go:embed templates/message.tpl
var messageTpl string

func GetTableForMessage(m kit.Message, p kit.Parameters) string {
	type Data struct {
		Message    kit.Message
		Parameters kit.Parameters
	}

	buf := bytes.NewBuffer([]byte{})
	t, err := template.New("message.tpl").
		Funcs(template.FuncMap{
			"GetFieldType": GetFieldType,
			"BoolIcon":     BoolIcon,
			"GetComments":  kit.GetComments,
		}).
		Parse(messageTpl)
	if err != nil {
		panic(err)
	}

	err = t.Execute(buf, Data{m, p})
	if err != nil {
		panic(err)
	}

	return buf.String()
}

//go:embed templates/enum.tpl
var enumTpl string

func GetTableForEnum(e kit.Enum, p kit.Parameters) string {
	type Data struct {
		Enum       kit.Enum
		Parameters kit.Parameters
	}

	buf := bytes.NewBuffer([]byte{})
	t, err := template.New("enum.tpl").
		Funcs(template.FuncMap{
			"GetComments": kit.GetComments,
		}).
		Parse(enumTpl)
	if err != nil {
		panic(err)
	}

	err = t.Execute(buf, Data{e, p})
	if err != nil {
		panic(err)
	}

	return buf.String()
}

//go:embed templates/service.tpl
var serviceTpl string

func GetTableForService(s kit.Service, p kit.Parameters) string {
	type Data struct {
		Service    kit.Service
		Parameters kit.Parameters
	}

	for _, method := range s.Methods {
		log.Printf("%v", method)
	}

	buf := bytes.NewBuffer([]byte{})
	t, err := template.New("service.tpl").
		Funcs(template.FuncMap{
			"GetComments": kit.GetComments,
		}).
		Parse(serviceTpl)
	if err != nil {
		panic(err)
	}

	err = t.Execute(buf, Data{s, p})
	if err != nil {
		panic(err)
	}

	return buf.String()

	var base strings.Builder
	// base.WriteString("[#" + strings.ToLower(s.serviceName) + "_" + strings.ToLower(s.name) + "]\n")
	base.WriteString("[[" + s.Name + "]]")
	base.WriteString("=== " + s.Name)

	base.WriteString("\n" + s.LeadingComments)
	base.WriteString("\n" + s.TrailingComments)

	// table overview of the service
	// TODO: include if it has REST or not
	base.WriteString("\n")
	base.WriteString(`[cols=">1,<3"]`)
	base.WriteString("\n")
	base.WriteString("\n|===")
	base.WriteString("\n| Name | " + s.Name)
	for _, method := range s.Methods {
		base.WriteString(GetMethodType(method, p))
	}

	/*
		base.WriteString("\n| Request Type | proto2asciidoc:message[" + s.requestsType + "]")
		base.WriteString("\n| Streaming Request | " + boolIcon(s.requestsStream))
		base.WriteString("\n| Return Type | proto2asciidoc:message[" + s.returnsType + "]")
		base.WriteString("\n| Streaming Return | " + boolIcon(s.returnsStream))
		base.WriteString("\n| REST Support | " + boolIcon(s.rest != nil))
		if s.rest != nil {
			base.WriteString("\n| REST Method | " + strings.ToUpper(s.rest.method))
			base.WriteString("\n| REST URL | `" + s.rest.url + "`")
			if s.rest.body != `` {
				bodyOutput := "`" + s.rest.body + "`"
				base.WriteString("\n| REST Body | " + bodyOutput)
			}

		}
	*/
	base.WriteString("\n|===")
	base.WriteString("\n")

	return base.String()
}

func GetMethodType(m kit.Method, p kit.Parameters) string {
	return fmt.Sprintf("\n%v\n", m)
	// if f.TypeName != nil && f.Type != nil {
	// 	fqtn := strings.Split(*f.TypeName, ".")
	// 	if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
	// 		if val, ok := p["extension"]; ok && val == "on" {
	// 			return fmt.Sprintf("proto2asciidoc:message[%s]", fqtn[len(fqtn)-1])
	// 		}
	// 		return fmt.Sprintf("<<%s_message>>", fqtn[len(fqtn)-1])
	// 	}
	// 	if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM {
	// 		if val, ok := p["extension"]; ok && val == "on" {
	// 			return fmt.Sprintf("proto2asciidoc:enum[%s]", fqtn[len(fqtn)-1])
	// 		}
	// 		return fmt.Sprintf("<<%s_message>>", fqtn[len(fqtn)-1])
	// 	}
	// }
	// return strings.ToLower(strings.TrimLeft(f.Type.String(), "TYPE_"))
	// return ""
}
