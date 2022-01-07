package formatter

import (
	"github.com/productsupcom/protoc-gen-kit/kit"
	"google.golang.org/protobuf/types/pluginpb"
)

type Formatter interface {
	GetFile(desc kit.Desc) *pluginpb.CodeGeneratorResponse_File
}
