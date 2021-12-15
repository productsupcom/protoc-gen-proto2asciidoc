package formatter

import (
	"github.com/productsupcom/proto2asciidoc/kit"
	"google.golang.org/protobuf/types/pluginpb"
)

type Formatter interface {
	GetFiles(desc kit.Desc) []*pluginpb.CodeGeneratorResponse_File
}
