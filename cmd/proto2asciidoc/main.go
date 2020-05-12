package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"github.com/spf13/pflag"

	"github.com/productsupcom/proto2asciidoc/cmd/proto2asciidoc/docs"
)

var (
	flags      *pflag.FlagSet
	sourceFile string
	outFile    string
	overwrite  bool
	buf        bytes.Buffer
	noheader   bool

	apidocs     bool
	apidir      string
	samplefiles []string

	docPaths docs.DocPaths
	samples  []docs.SampleFile

	messages docs.ProtoCollectionStruct
	enums    docs.ProtoCollectionStruct
	services docs.ProtoServiceOverview
)

func init() {
	/*
		Used by documentation for the manpage
		tag::options[]
		*--source string*
			Source file to parse into AsciiDoc, recommended is to set the absolute path.

		*--out*
			File to write to, if left empty writes to stdout

		*--overwrite, -f*
			Overwrite the existing out file

		*--no-header*
			Do not set a document header and ToC

		*--api-dir*
			Relative path from the out to the api dir. E.g. docs/generated/api.adoc is the out,
			the api dir is docs/api/
			then set --api-dir ../api

		This path will be used to set the includes for asciidoc

		*--api-docs*
			Generate a full API documentation, including files from the out file relative dir
			../api/SERVICE/endpoint.adoc.

		This will include every file in the appropriate location, so e.g. if you have declared Service Foo with endpoint
		Do() then it will check for these two files:

		/docs/api/foo.adoc
		/docs/api/foo/do.adoc

		Other files it will check for are:

		/docs/about.adoc
		/docs/examples.adoc
		/docs/api/errors.adoc

		Do not set if you only want Messages and/or Enums
		end::options[]
	*/
	flags = pflag.NewFlagSet("proto2asciidoc", pflag.ContinueOnError)
	flags.StringVar(&sourceFile, "source", "", "Source Protobuf file to parse into AsciiDoc, recommended is to set the absolute path.")
	flags.StringVar(&outFile, "out", "", "File to write to, if left empty writes to stdout")
	flags.BoolVarP(&overwrite, "overwrite", "f", false, "Overwrite the existing out file")
	flags.BoolVar(&noheader, "no-header", false, "Do not set a document header and ToC")
	flags.BoolVar(&apidocs, "api-docs", false, `Generate a full API documentation, including files from the out file relative dir
	../api/SERVICE/endpoint.adoc.

This will include every file in the appropriate location, so e.g. if you have declared Service Foo with endpoint
Do() then it will check for these two files:

/docs/api/foo.adoc
/docs/api/foo/do.adoc

It will also check for command documentation

/docs/cmd/foo.adoc

Other files it will check for are:

/docs/about.adoc
/docs/examples.adoc
/docs/api/errors.adoc

Do not set if you only want Messages and/or Enums`)
	flags.StringSliceVar(&samplefiles, "sample-files", []string{}, "List of files to use as sample variables. api/foo_samples.adoc becomes :foo_samples:")
	flags.StringVar(&apidir, "api-dir", "../api", `Relative (or absolute) path from the out to the api dir. E.g. docs/generated/api.adoc is the out,
the api dir is docs/api/
then set --api-dir ../api`)
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err != pflag.ErrHelp {
			fmt.Fprint(os.Stderr, err.Error()+"\n")
			flags.PrintDefaults()
		}
		os.Exit(100)
	}

	if sourceFile == "" {
		fmt.Fprint(os.Stderr, "Sourcefile must be set\n")
		flags.PrintDefaults()
		os.Exit(100)
	}

	reader, err := os.Open(sourceFile)
	if err != nil {
		exitError("Could not open source file", err)
	}

	docPaths = docs.DocPaths{
		Sourcefile:  sourceFile,
		Destination: outFile,
		ApiDir:      apidir,
	}

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		exitError("Could not parse", err)
	}

	if !apidocs {
		if !noheader {
			buf.WriteString("= Protobuffer Declarations\n")
			buf.WriteString(":toc: left\n")
			buf.WriteString(":toclevels: 4\n")
		}
	} else {
		buf.WriteString(`
= {project-name} Documentation
:doctype: book
:toc: left
:toclevels: 3
:icons: font
{project-author} <{project-repo}>
Version {version}

:apidoc: true
:sectnums:
:sectnumlevels: 2
`)
		for _, file := range samplefiles {
			_, basename := path.Split(file)

			samples = append(samples, docs.SampleFile{
				Name: basename,
				Path: file,
			})
			if strings.Contains(file, "adoc") {
				name := strings.TrimSuffix(basename, filepath.Ext(basename))
				if name != `` {
					buf.WriteString(fmt.Sprintf(":%s: %s\n", name, file))
				}
			}
		}
	}

	buf.WriteString(":true-icon: ✅\n")
	buf.WriteString(":false-icon: ❌\n")
	buf.WriteString("\n// WARNING \n// NEW THIS FILE IS GENERATED. DO NOT EDIT.\n\n")
	proto.Walk(definition,
		proto.WithService(handleService),
		proto.WithMessage(handleMessage),
		proto.WithEnum(handleEnum),
	)

	services.SetMessages(messages)
	services.SetEnums(enums)

	if apidocs {
		var preface bytes.Buffer

		preface.WriteString("// start included about.adoc (if found)\n")
		if file := docPaths.GetFilepathFor("../about.adoc"); file != "" {
			preface.WriteString("include::" + file + "[leveloffset=+1]\n")
		}
		preface.WriteString("// end included about.adoc\n\n")
		preface.WriteString("// start included examples.adoc (if found)\n")
		if file := docPaths.GetFilepathFor("../examples.adoc"); file != "" {
			preface.WriteString("include::" + file + "[leveloffset=+1]\n")
		}
		preface.WriteString("// end included examples.adoc\n\n")

		preface.WriteString("\n// start variables for the REST API endpoints\n")
		for _, service := range services.Collection() {
			service := service.(*docs.ProtoService)
			for endpoint, url := range service.GetRESTEndpoints() {
				preface.WriteString(fmt.Sprintf(":%s_%s_rest: %s\n",
					service.GetName(),
					endpoint,
					url,
				))
			}
			preface.WriteString("// end variables for the REST API endpoints\n\n")
		}

		// get the cmd docs
		files := docPaths.GetFilesInDir("../cmd")
		preface.WriteString("// start included files from the /cmd directory (if found any)\n")
		for _, file := range files {
			preface.WriteString("include::" + file + "[leveloffset=+1]\n")
		}
		preface.WriteString("// end included files from the /cmd directory\n\n")

		if preface.Len() != 0 {
			buf.Write(preface.Bytes())
		}

		buf.WriteString(services.GetOutput())

		if file := docPaths.GetFilepathFor("errors.adoc"); file != "" {
			buf.WriteString("include::" + file + "[leveloffset=+1]\n")
		}
		buf.WriteString("== Protobuffer Declarations\n")
		buf.WriteString("=== Protobuf Enums\n")
		buf.WriteString("\n:leveloffset: +2\n")
	}

	enums.Sort()
	var enumNames []string
	for _, enum := range enums.Collection() {
		enumNames = append(enumNames, enum.GetName())
	}
	messages.Sort()
	var messageNames []string
	for _, message := range messages.Collection() {
		messageNames = append(messageNames, message.GetName())
	}

	for _, enum := range enums.Collection() {
		enum.SetEnumNames(enumNames)
		enum.SetMessageNames(messageNames)
		buf.WriteString("[#" + strings.ToLower(enum.GetName()) + "_enum]\n")
		buf.WriteString(docs.GetOutput(enum))
	}

	if apidocs {
		buf.WriteString("\n:leveloffset: -2\n")
		buf.WriteString("=== Protobuf Messages\n")
		buf.WriteString("\n:leveloffset: +2\n")
	}

	for _, message := range messages.Collection() {
		message.SetEnumNames(enumNames)
		message.SetMessageNames(messageNames)
		buf.WriteString("[#" + strings.ToLower(message.GetName()) + "_message]\n")
		buf.WriteString(docs.GetOutput(message))
	}
	if apidocs {
		buf.WriteString("\n:leveloffset: -2\n")
	}

	if apidocs {
		dir, _ := path.Split(outFile)
		dir += "api/*.adoc"
		files, err := filepath.Glob(dir)
		if err == nil {

			buf.WriteString("=== Imported Protobuf Declarations\n")
			for _, file := range files {
				_, filename := path.Split(file)
				buf.WriteString("include::api/" + filename + "[leveloffset=+2]\n")

			}
		}
	}

	if outFile != `` {
		if stat, _ := os.Stat(outFile); stat != nil {
			if !overwrite {
				exitError("File already exists", nil)
			}
			if err = os.Remove(outFile); err != nil {
				exitError("Could not delete file", err)
			}
		}
		o, err := os.Create(outFile)
		if err != nil {
			exitError("Could not create file for writing: "+outFile, err)
		}

		_, err = o.Write(buf.Bytes())
		if err != nil {
			exitError("Could not write string to file: "+outFile, err)
		}
	} else {
		fmt.Println(buf.String())
	}

	os.Exit(0)
}

func handleService(s *proto.Service) {
	service := docs.NewProtoService(s, docPaths, samples)
	for _, e := range s.Elements {
		if r, ok := e.(*proto.RPC); ok {
			sf := docs.NewProtoServiceField(r)
			service.Append(sf)
		}
	}
	services.Append(&service)
}

func handleEnum(m *proto.Enum) {
	enum := docs.NewProtoEnum(m, docPaths, samples)

	for _, e := range m.Elements {
		if c, ok := e.(*proto.EnumField); ok {
			f := docs.NewProtoEnumField(c)
			enum.Append(f)
		}
	}

	enums.Append(&enum)
}

func handleMessage(m *proto.Message) {
	message := docs.NewProtoMessage(m, docPaths, samples)

	for _, e := range m.Elements {
		if c, ok := e.(*proto.NormalField); ok {
			f := docs.NewProtoMessageField(c)
			message.Append(f)
		}
		if c, ok := e.(*proto.MapField); ok {
			f := docs.NewProtoMessageMapField(c)
			message.Append(f)
		}
	}

	messages.Append(&message)
}

func exitError(reason string, err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, reason+": "+err.Error()+"\n")
	} else {
		fmt.Fprint(os.Stderr, reason+"\n")
	}
	os.Exit(1)
}
