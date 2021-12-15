package kit

import (
	"sort"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type Parameters map[string]string

func ProcessParameters(request *pluginpb.CodeGeneratorRequest) Parameters {
	p := Parameters{}
	for _, e := range strings.Split(request.GetParameter(), ",") {
		kv := strings.Split(e, "=")
		if len(kv) == 2 {
			p[kv[0]] = kv[1]
			continue
		}
		p[kv[0]] = ""
	}
	return p
}

type Desc struct {
	Parameters Parameters
	Messages   []Message
	Enums      []Enum
	Services   []Service
}

func DescFromRequest(request *pluginpb.CodeGeneratorRequest) Desc {
	desc := Desc{}

	desc.Parameters = ProcessParameters(request)

	for _, want := range request.FileToGenerate {
		for _, protofile := range request.ProtoFile {
			if *protofile.Name == want {
				desc.Messages = MessagesFromFileDesc(protofile)
				desc.Enums = EnumsFromFileDesc(protofile)
				desc.Services = ServicesFromFileDesc(protofile)
			}
		}
	}

	if val, ok := desc.Parameters["sorted"]; ok && val == "on" {
		sort.SliceStable(desc.Messages, func(i, j int) bool {
			return desc.Messages[i].Name < desc.Messages[j].Name
		})

		sort.SliceStable(desc.Enums, func(i, j int) bool {
			return desc.Enums[i].Name < desc.Enums[j].Name
		})

		sort.SliceStable(desc.Services, func(i, j int) bool {
			return desc.Services[i].Name < desc.Services[j].Name
		})
	}

	return desc
}

type Message struct {
	Name string
	Comments
	Fields []Field
}

type Field struct {
	*descriptorpb.FieldDescriptorProto
	Comments
}

func (f *Field) Repeated() bool {
	return f.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
}

func GetComments(c Comments) string {
	return c.GetComments()
}

func (f *Field) HasComments() bool {
	return len(strings.Trim(f.LeadingComments, "\n")) > 0 || len(strings.Trim(f.TrailingComments, "\n")) > 0
}

func MessagesFromFileDesc(protofile *descriptorpb.FileDescriptorProto) []Message {
	var msgs []Message

	for _, mt := range protofile.MessageType {
		msg := Message{
			Name: *mt.Name,
		}

		for _, f := range mt.Field {
			field := Field{
				FieldDescriptorProto: f,
			}

			msg.Fields = append(msg.Fields, field)
		}

		msgs = append(msgs, msg)
	}

	for _, loc := range protofile.SourceCodeInfo.Location {
		if len(loc.Path) < 2 {
			continue
		}
		if loc.Path[0] == 4 {
			// this is the message
			if len(loc.Path) == 2 {
				if loc.LeadingComments != nil {
					msgs[loc.Path[1]].LeadingComments = cleanComment(*loc.LeadingComments)
				}
				if loc.TrailingComments != nil {
					msgs[loc.Path[1]].TrailingComments = cleanComment(*loc.TrailingComments)
				}
			}
			// this is the field
			if len(loc.Path) == 4 {
				if loc.LeadingComments != nil {
					msgs[loc.Path[1]].Fields[loc.Path[3]].LeadingComments = cleanComment(*loc.LeadingComments)
				}
				if loc.TrailingComments != nil {
					msgs[loc.Path[1]].Fields[loc.Path[3]].TrailingComments = cleanComment(*loc.TrailingComments)
				}
			}
		}
	}

	return msgs
}

type Enum struct {
	Name string
	Comments
	Values []Value
}

type Value struct {
	*descriptorpb.EnumValueDescriptorProto
	Comments
}

func EnumsFromFileDesc(protofile *descriptorpb.FileDescriptorProto) []Enum {
	var enums []Enum

	for _, et := range protofile.EnumType {
		enum := Enum{
			Name: *et.Name,
		}

		for _, v := range et.Value {
			value := Value{
				EnumValueDescriptorProto: v,
			}

			enum.Values = append(enum.Values, value)
		}

		enums = append(enums, enum)
	}

	for _, loc := range protofile.SourceCodeInfo.Location {
		if len(loc.Path) < 2 {
			continue
		}
		if loc.Path[0] == 5 {
			// // this is the enum declaration
			if len(loc.Path) == 2 {
				if loc.LeadingComments != nil {
					enums[loc.Path[1]].LeadingComments = cleanComment(*loc.LeadingComments)
				}
				if loc.TrailingComments != nil {
					enums[loc.Path[1]].TrailingComments = cleanComment(*loc.TrailingComments)
				}
			}
			// this is the field
			if len(loc.Path) == 4 {
				if loc.LeadingComments != nil {
					enums[loc.Path[1]].Values[loc.Path[3]].LeadingComments = cleanComment(*loc.LeadingComments)
				}
				if loc.TrailingComments != nil {
					enums[loc.Path[1]].Values[loc.Path[3]].TrailingComments = cleanComment(*loc.TrailingComments)
				}
			}
		}
	}

	return enums
}

type Service struct {
	Name    string
	Methods []Method
	Comments
}

type Method struct {
	*descriptorpb.MethodDescriptorProto
	Comments
}

// func (m *Method) SupportsREST() bool {
// 	if m.Options ==
// }

func ServicesFromFileDesc(protofile *descriptorpb.FileDescriptorProto) []Service {
	var services []Service

	for _, srv := range protofile.Service {
		service := Service{
			Name: *srv.Name,
		}

		for _, m := range srv.Method {
			method := Method{
				MethodDescriptorProto: m,
			}

			service.Methods = append(service.Methods, method)
		}

		services = append(services, service)
	}

	for _, loc := range protofile.SourceCodeInfo.Location {
		if len(loc.Path) < 2 {
			continue
		}
		if loc.Path[0] == 6 {
			// log.Printf("Service found: \t%v\n", loc)
			// // this is the enum declaration
			if len(loc.Path) == 2 {
				if loc.LeadingComments != nil {
					services[loc.Path[1]].LeadingComments = cleanComment(*loc.LeadingComments)
				}
				if loc.TrailingComments != nil {
					services[loc.Path[1]].TrailingComments = cleanComment(*loc.TrailingComments)
				}
			}
			// this is the field
			if len(loc.Path) == 4 {
				if loc.LeadingComments != nil {
					services[loc.Path[1]].Methods[loc.Path[3]].LeadingComments = cleanComment(*loc.LeadingComments)
				}
				if loc.TrailingComments != nil {
					services[loc.Path[1]].Methods[loc.Path[3]].TrailingComments = cleanComment(*loc.TrailingComments)
				}
			}
		}
	}

	return services
}

func (c *Comments) GetComments() string {
	var buf []string
	if c.LeadingComments != "" {
		buf = append(buf, strings.Trim(c.LeadingComments, " "))
	}
	if c.TrailingComments != "" {
		buf = append(buf, strings.Trim(c.TrailingComments, " "))
	}
	return strings.Join(buf, "\n")
}

type Comments struct {
	LeadingComments  string
	TrailingComments string
}

func cleanComment(s string) string {
	var out []string
	buf := strings.Split(s, "\n")
	for _, line := range buf {
		if len(line) == 0 {
			continue
		}
		if !strings.Contains(line, "tag::") && !strings.Contains(line, "end::") {
			out = append(out, line)
		}
	}

	return strings.Join(out, "\n")
}
