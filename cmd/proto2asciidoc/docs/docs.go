package docs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/emicklei/proto"
)

type SampleFile struct {
	Name     string
	Path     string
	Relative string
}

type DocPaths struct {
	Sourcefile  string
	Destination string
	ApiDir      string
}

type docBase interface {
	setComments(*proto.Comment)
	getComments() string
	getCollapsible() bool
	GetPaths() DocPaths
	SetMessageNames([]string)
	SetEnumNames([]string)
}
type docBaseStruct struct {
	name        string
	comments    []string
	typeName    string
	paths       DocPaths
	collapsible bool
	samples     []SampleFile
	msgNames    []string
	enumNames   []string
}

func (d *docBaseStruct) SetMessageNames(names []string) {
	d.msgNames = names
}

func (d *docBaseStruct) SetEnumNames(names []string) {
	d.enumNames = names
}

func (d *docBaseStruct) GetMessageNames() []string {
	return d.msgNames
}

func (d *docBaseStruct) GetEnumNames() []string {
	return d.enumNames
}

func (d *docBaseStruct) getCollapsible() bool {
	return d.collapsible
}

func (d *docBaseStruct) GetPaths() DocPaths {
	return d.paths
}

type ProtoCollection interface {
	getProtoByName(string) docSection
	Append(docSection)
	Collection() []docSection
}

type ProtoCollectionStruct struct {
	collection []docSection
}

func (c *ProtoCollectionStruct) Append(doc docSection) {
	c.collection = append(c.collection, doc)
}

func (c *ProtoCollectionStruct) Collection() []docSection {
	return c.collection
}

func (c *ProtoCollectionStruct) Sort() {
	sort.SliceStable(c.collection, func(i, j int) bool {
		return c.collection[i].GetName() < c.collection[j].GetName()
	})
}

func (c *ProtoCollectionStruct) getProtoByName(name string) *docSection {
	for _, doc := range c.collection {
		if doc.GetName() == name {
			return &doc
		}
	}
	return nil
}

func (p *docBaseStruct) getSampleByName(name string) *SampleFile {
	for _, sample := range p.samples {
		if sample.Name == name {
			return &sample
		}
	}

	return nil
}

type ProtoServiceOverview struct {
	ProtoCollectionStruct
	messages ProtoCollectionStruct
	enums    ProtoCollectionStruct
}

func (o *ProtoServiceOverview) SetMessages(messages ProtoCollectionStruct) {
	o.messages = messages
}

func (o *ProtoServiceOverview) SetEnums(enums ProtoCollectionStruct) {
	o.enums = enums
}

type docSection interface {
	docBase
	GetName() string
	GetAnchor() string
	getTable() string
	getTitle() string
	getSourceFile() string
}

type docSectionField interface {
	docBase
	getRow(string, []string) string
	getType([]string) string
}

type ProtoService struct {
	docBaseStruct
	fields []protoServiceField
}

type ProtoRestService struct {
	method string
	url    string
	body   string
}

type protoServiceField struct {
	docBaseStruct
	requestsType   string
	returnsType    string
	returnsStream  bool
	requestsStream bool
	rest           *ProtoRestService
	serviceName    string
}

func (p *protoServiceField) getTable() string {
	return ""
}

type ProtoMessage struct {
	docBaseStruct
	fields []protoMessageField
}

type protoMessageField struct {
	docBaseStruct
	repeated    bool
	messageType string
	position    int
}

type ProtoEnum struct {
	docBaseStruct
	fields []protoEnumField
}

type protoEnumField struct {
	docBaseStruct
	enum int
}

func NewProtoService(s *proto.Service, paths DocPaths, samples []SampleFile) ProtoService {
	service := ProtoService{}
	service.name = s.Name
	service.setComments(s.Comment)
	service.typeName = "Service"
	service.paths = paths
	for _, sample := range samples {
		sample.ProcessPath(paths)
		service.samples = append(service.samples, sample)
	}
	return service
}

func NewProtoMessage(m *proto.Message, paths DocPaths, samples []SampleFile) ProtoMessage {
	msg := ProtoMessage{}
	msg.name = m.Name
	msg.setComments(m.Comment)
	msg.typeName = "Message"
	msg.paths = paths
	for _, sample := range samples {
		sample.ProcessPath(paths)
		msg.samples = append(msg.samples, sample)
	}
	return msg
}

func NewProtoEnum(m *proto.Enum, paths DocPaths, samples []SampleFile) ProtoEnum {
	en := ProtoEnum{}
	en.name = m.Name
	en.setComments(m.Comment)
	en.typeName = "Enum"
	en.paths = paths
	for _, sample := range samples {
		sample.ProcessPath(paths)
		en.samples = append(en.samples, sample)
	}
	return en
}

func (s *SampleFile) ProcessPath(paths DocPaths) {
	if strings.Contains(s.Path, "generated") {
		cur := strings.SplitAfter(s.Path, "generated")
		if len(cur) > 1 {
			s.Relative = "../generated" + cur[1]
		}
		return
	}
	if strings.Contains(s.Path, "api/") {
		apidir := path.Base(paths.ApiDir)
		if apidir != `.` {
			cur := strings.SplitAfter(s.Path, apidir)
			if len(cur) > 1 {
				s.Relative = paths.ApiDir + cur[1]
			}
		}
	}
}

func (m *ProtoEnum) Append(field protoEnumField) {
	field.paths = m.paths
	m.fields = append(m.fields, field)
}

func (m *ProtoMessage) Append(field protoMessageField) {
	field.paths = m.paths
	m.fields = append(m.fields, field)
}

func (s *ProtoService) Append(field protoServiceField) {
	field.paths = s.paths
	field.serviceName = s.name
	s.fields = append(s.fields, field)
}

func (s *ProtoService) GetRESTEndpoints() map[string]string {
	endpoints := make(map[string]string)
	for _, field := range s.fields {
		if field.rest != nil {
			endpoints[field.name] = field.rest.url
		}
	}
	return endpoints
}

func NewProtoServiceField(r *proto.RPC) protoServiceField {
	sf := protoServiceField{}
	sf.name = r.Name
	sf.typeName = "Endpoint"
	sf.requestsStream = r.StreamsRequest
	sf.returnsStream = r.StreamsReturns
	sf.requestsType = r.RequestType
	sf.returnsType = r.ReturnsType
	sf.setComments(r.Comment)

	for _, e := range r.Elements {
		if o, ok := e.(*proto.Option); ok {
			rest := &ProtoRestService{}
			rest.method = o.Constant.OrderedMap[0].Name
			rest.url = o.Constant.OrderedMap[0].Source
			if len(o.Constant.OrderedMap) > 1 {
				rest.body = o.Constant.OrderedMap[1].Source
			}

			sf.rest = rest
		}
	}

	return sf
}

func NewProtoEnumField(f *proto.EnumField) protoEnumField {
	en := protoEnumField{}
	en.name = f.Name
	en.enum = f.Integer
	en.setComments(f.InlineComment)

	return en
}

func NewProtoMessageField(m *proto.NormalField) protoMessageField {
	f := protoMessageField{}
	f.name = m.Name
	f.messageType = m.Field.Type
	f.position = m.Field.Sequence
	f.repeated = m.Repeated
	f.setComments(m.Field.InlineComment)

	return f
}

func NewProtoMessageMapField(m *proto.MapField) protoMessageField {
	f := protoMessageField{}
	fieldtype := fmt.Sprintf("map[%s]%s", m.KeyType, m.Field.Type)
	f.name = m.Name
	f.messageType = fieldtype
	f.position = m.Field.Sequence
	f.repeated = false
	f.setComments(m.Field.InlineComment)

	return f
}

func getSectionHeader(d docSection) string {
	var base strings.Builder
	base.WriteString("== " + d.getTitle() + "\n")

	base.WriteString("// tag::" + d.GetName() + "[]\n")
	base.WriteString(d.getComments())

	return base.String()
}

func getProtobufSource(d docSection) string {
	var base strings.Builder

	base.WriteString("\n.Protobuf Source")
	base.WriteString("\n[%collapsible]\n")
	base.WriteString("====\n")
	base.WriteString("[source,protobuf]\n")
	base.WriteString("----\n")
	base.WriteString("include::" + d.getSourceFile() + "[tag=" + d.GetName() + "]")
	base.WriteString("\n----\n")
	base.WriteString("====\n")

	return base.String()
}

func getCodeExample(message *ProtoMessage) string {
	sample := message.getSampleByName("message_samples.adoc")
	if sample != nil {
		found := false
		tagName := message.name + "Message"
		if _, err := os.Stat(sample.Path); err == nil {
			data, err := ioutil.ReadFile(sample.Path)
			if err == nil {
				var contents []string
				contents = append(contents, strings.Split(string(data), "\n")...)
				for _, str := range contents {
					if strings.Contains(str, tagName) {
						found = true
					}
				}
			}
		}
		if found {
			return "\ninclude::" + sample.Relative + "[tag=" + tagName + ", leveloffset=+1]\n"
		}
	}

	return ``
}

func GetOutput(d docSection) string {
	var base strings.Builder

	base.WriteString(getSectionHeader(d))
	base.WriteString(d.getTable())
	base.WriteString(getProtobufSource(d))
	if message, ok := d.(*ProtoMessage); ok {
		base.WriteString(getCodeExample(message))
	}
	base.WriteString("// end::" + d.GetName() + "[]\n")
	base.WriteString("\n")

	return base.String()
}

func (o *ProtoServiceOverview) GetOutput() string {
	var base strings.Builder

	// base.WriteString("= Overview\n")

	for _, service := range o.Collection() {
		service := service.(*ProtoService)

		base.WriteString("[#" + strings.ToLower(service.name) + "_service]\n")
		base.WriteString(getSectionHeader(service))

		// check if the service has it's own written manual, for intros etc.
		file := service.paths.GetFilepathFor(strings.ToLower(service.name) + ".adoc")
		if file != "" {
			base.WriteString("include::" + file + "[leveloffset=+2]\n")
		}

		for _, endpoint := range service.fields {
			base.WriteString(endpoint.getSection())
		}
	}
	// base.WriteString(getSectionHeader(service))

	return base.String()
}

func boolIcon(val bool) string {
	if val {
		return "{true-icon}"
	}
	return "{false-icon}"
}

func (s *protoServiceField) getSection() string {
	var base strings.Builder

	base.WriteString("<<<\n")
	base.WriteString("[#" + strings.ToLower(s.serviceName) + "_" + strings.ToLower(s.name) + "]\n")
	base.WriteString("=== " + s.getTitle())

	base.WriteString("\n" + s.getComments())

	// table overview of the service
	// TODO: include if it has REST or not
	base.WriteString("\n")
	base.WriteString(`[cols=">1,<3"]`)
	base.WriteString("\n")
	base.WriteString("\n|===")
	base.WriteString("\n| Name | " + s.name)
	base.WriteString("\n| Request Type | distrib:message[" + s.requestsType + "]")
	base.WriteString("\n| Streaming Request | " + boolIcon(s.requestsStream))
	base.WriteString("\n| Return Type | distrib:message[" + s.returnsType + "]")
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
	base.WriteString("\n|===")
	base.WriteString("\n")

	base.WriteString(getProtobufSource(s))

	file := s.paths.GetFilepathFor(strings.ToLower(s.serviceName) + "/" + strings.ToLower(s.name) + ".adoc")
	if file != "" {
		base.WriteString("include::" + file + "[leveloffset=+2]")
	}

	base.WriteString("\n")

	return base.String()
}

func (p DocPaths) GetFilepathFor(filename string) string {
	dir, _ := path.Split(p.Destination)
	relative := p.ApiDir + "/" + filename
	dir += relative
	if _, err := os.Stat(dir); err == nil {
		// check if the relative path got bananas, asciidoc doesn't like stuff like
		// ../api/../file.adoc
		// go doesn't care :)

		if strings.Contains(relative, p.ApiDir+"/../") {
			// in this case the filename was alright
			return filename
		}
		return relative

	}
	return ""
}

func (d *docBaseStruct) GetName() string {
	return d.name
}

func (d *docBaseStruct) GetAnchor() string {
	return strings.ToLower(d.name)
}

func (d *docBaseStruct) getSourceFile() string {
	return d.paths.Sourcefile
}

func (m *ProtoMessage) getTable() string {
	var table strings.Builder

	table.WriteString("\n." + m.GetName() + " Overview")
	if m.collapsible {
		table.WriteString("\n[%collapsible]")
		table.WriteString("\n====")
	}
	table.WriteString("\n" + `[cols="2,2,1,1", options="header"]`)
	table.WriteString("\n|===\n")
	table.WriteString("| Name | Type | Repeated | Sequence")
	for _, f := range m.fields {
		f.SetEnumNames(m.GetEnumNames())
		f.SetMessageNames(m.GetMessageNames())
		table.WriteString(f.getRow(m.GetName()))
	}
	if len(m.fields) == 0 {
		table.WriteString("\n")
	}
	table.WriteString("\n|===\n")
	if m.collapsible {
		table.WriteString("====\n")
	}

	return table.String()
}

func (f *protoMessageField) getRow(name string) string {
	var row strings.Builder
	row.WriteString("\n|[[" + strings.ToLower(name) + "_" + f.name + "]]" + f.name)
	row.WriteString("\n|" + f.getType())
	row.WriteString("\n|" + boolIcon(f.repeated))
	row.WriteString("\n|" + strconv.Itoa(f.position))
	if len(strings.Trim(f.getComments(), "\n ")) > 0 {
		row.WriteString("\n\n|  \n3+|" + f.getComments())
	}

	return row.String()
}

func (m *ProtoEnum) getTable() string {
	var table strings.Builder

	table.WriteString("\n." + m.GetName() + " Overview")
	if m.collapsible {
		table.WriteString("\n[%collapsible]")
		table.WriteString("\n====")
	}
	table.WriteString("\n" + `[cols="2,1,3", options="header"]`)
	table.WriteString("\n|===\n")
	table.WriteString("| Name | Integer | Comment")
	for _, f := range m.fields {
		table.WriteString(f.getRow(m.GetName()))
	}
	if len(m.fields) == 0 {
		table.WriteString("\n")
	}
	table.WriteString("\n|===\n")
	if m.collapsible {
		table.WriteString("====\n")
	}

	return table.String()
}

func (m *ProtoService) getTable() string {
	var table strings.Builder

	table.WriteString("\n." + m.GetName() + " Overview")
	table.WriteString("\n[%collapsible]")
	table.WriteString("\n====")
	// table.WriteString("\n" + `[cols="2,2,1,2,1,3"]`)
	table.WriteString("\n|===\n")
	table.WriteString("| Name | Requests | Streams | Returns | Streams | Comment")
	for _, f := range m.fields {
		table.WriteString(f.getRow(m.GetName()))
	}
	if len(m.fields) == 0 {
		table.WriteString("\n")
	}
	table.WriteString("\n|===")
	table.WriteString("\n====\n")

	return table.String()
}

func (f *protoServiceField) getRow(name string) string {
	var row strings.Builder
	row.WriteString("\n|" + f.name)
	row.WriteString("\n|" + f.requestsType)
	row.WriteString("\n|" + fmt.Sprintf("%v", f.requestsStream))
	row.WriteString("\n|" + f.returnsType)
	row.WriteString("\n|" + fmt.Sprintf("%v", f.requestsStream))
	row.WriteString("\n|" + f.getComments())

	return row.String()
}

func (f *protoMessageField) getType() string {
	if f.messageType == `` {
		return ``
	}
	for _, enum := range f.GetEnumNames() {
		if f.messageType == enum {
			return fmt.Sprintf("distrib:enum[%s]", f.messageType)
		}
	}
	for _, msg := range f.GetMessageNames() {
		if f.messageType == msg {
			return fmt.Sprintf("distrib:message[%s]", f.messageType)
		}
	}
	if f.messageType != "string" &&
		!strings.Contains(f.messageType, "map[") &&
		!strings.Contains(f.messageType, "int") &&
		!strings.Contains(f.messageType, "bool") {
		if strings.Contains(f.messageType, ".") {
			return fmt.Sprintf("<<%s>>", strings.ReplaceAll(f.messageType, ".", "_"))
		}
		return fmt.Sprintf("<<%s>>", f.messageType)
	}

	return f.messageType
}

func (f *protoEnumField) getRow(name string) string {
	var row strings.Builder
	row.WriteString("\n|[[" + name + "_" + f.name + "]]" + f.name)
	row.WriteString("\n|" + strconv.Itoa(f.enum))
	row.WriteString("\n|" + f.getComments())

	return row.String()
}

func (m *docBaseStruct) setComments(comment *proto.Comment) {
	if comment != nil {
		m.comments = comment.Lines
	}
}

func (d *docBaseStruct) getTitle() string {
	return fmt.Sprintf("%s %s", d.name, d.typeName)
}

func (d *docBaseStruct) getComments() string {
	var out strings.Builder
	for _, c := range d.comments {
		if !strings.Contains(c, "tag::") {
			b := c
			if strings.Contains(d.getSourceFile(), "google") {
				if b == " " {
					b = strings.Replace(b, " ", "----", 1)
				}
				b = strings.Replace(b, "# ", "=== ", 1)
				b = strings.Replace(b, "Example ", ".Example ", 1)
			}
			out.WriteString(strings.Trim(b, "\t *") + "\n")
		}
	}
	if len(d.comments) == 0 {
		out.WriteString("\n")
	}
	return out.String()
}
