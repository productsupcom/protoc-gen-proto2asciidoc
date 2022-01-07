package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/productsupcom/protoc-gen-kit/kit"
	"github.com/productsupcom/protoc-gen-proto2asciidoc/formatter"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	request  = &pluginpb.CodeGeneratorRequest{}
	response = &pluginpb.CodeGeneratorResponse{}

	writeWire = false
)

func main() {
	os.Stdout.Write(process(os.Stdin))
}

func process(r io.Reader) []byte {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	// used for the unit tests
	if writeWire {
		os.WriteFile("protowire", data, 0644)
		os.Exit(0)
	}

	if err := proto.Unmarshal(data, request); err != nil {
		panic(err)
	}

	kit.CleanCommentFn = formatter.CleanComment
	descs := kit.DescsFromRequest(request)

	var output formatter.Asciidoc
	for _, desc := range descs {
		response.File = append(response.File, output.GetFile(desc))
	}

	out, err := proto.Marshal(response)
	if err != nil {
		panic(err)
	}
	return out
}
