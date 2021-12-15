package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/productsupcom/proto2asciidoc/formatter"
	"github.com/productsupcom/proto2asciidoc/kit"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	request  = &pluginpb.CodeGeneratorRequest{}
	response = &pluginpb.CodeGeneratorResponse{}
	desc     kit.Desc

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

	desc = kit.DescFromRequest(request)

	var output formatter.Asciidoc
	response.File = output.GetFiles(desc)

	out, err := proto.Marshal(response)
	if err != nil {
		panic(err)
	}
	return out
}
