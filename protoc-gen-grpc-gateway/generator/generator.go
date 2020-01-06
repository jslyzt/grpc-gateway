package generator

import (
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/jslyzt/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
)

// Generator is an abstraction of code generators.
type Generator interface {
	Generate(targets []*descriptor.File) ([]*plugin.CodeGeneratorResponse_File, error) // generates output files from input .proto files
}
