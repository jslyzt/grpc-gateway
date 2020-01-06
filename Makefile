GO_PROTOBUF_REPO=github.com/golang/protobuf
GO_PLUGIN_PKG=$(GO_PROTOBUF_REPO)/protoc-gen-go
GOOGLEAPIS_DIR=third_party/googleapis
ADDITIONAL_GW_FLAGS=,grpc_api_configuration=examples/proto/examplepb/unannotated_echo_service.yaml
ADDITIONAL_SWG_FLAGS=,grpc_api_configuration=examples/proto/examplepb/unannotated_echo_service.yaml
PKGMAP=Mgoogle/protobuf/field_mask.proto=google.golang.org/genproto/protobuf/field_mask,Mgoogle/protobuf/descriptor.proto=$(GO_PLUGIN_PKG)/descriptor,Mexamples/proto/sub/message.proto=github.com/jslyzt/grpc-gateway/examples/proto/sub


.PHONY: gateway swagger install clean

gateway:
	go build protoc-gen-grpc-gateway -o bin/protoc-gen-grpc-gateway main.go

swagger:
	go build protoc-gen-swagger -o bin/protoc-gen-swagger main.go

install:
	cd protoc-gen-grpc-gateway; go install
	cd protoc-gen-swagger; go install

clean:
	rm -fr bin/*

.PHONY: emp_protoc examples

emp_protoc:
	protoc --plugin=protoc-gen-go -I. --go_out=$(PKGMAP),paths=source_relative:. internal/stream_chunk.proto
	protoc --plugin=protoc-gen-go -I. --go_out=$(PKGMAP),paths=source_relative:. protoc-gen-swagger/options/openapiv2.proto protoc-gen-swagger/options/annotations.proto
	protoc --plugin=protoc-gen-go -I. -I$(GOOGLEAPIS_DIR) --go_out=$(PKGMAP),plugins=grpc,paths=source_relative:. examples/proto/examplepb/*.proto
	protoc --plugin=bin/protoc-gen-grpc-gateway -I. -I$(GOOGLEAPIS_DIR) --grpc-gateway_out=logtostderr=true,allow_repeated_fields_in_body=true,$(PKGMAP)$(ADDITIONAL_GW_FLAGS):. examples/proto/examplepb/*.proto
	protoc --plugin=bin/protoc-gen-swagger -I. -I$(GOOGLEAPIS_DIR) --swagger_out=logtostderr=true,allow_repeated_fields_in_body=true,use_go_templates=true,$(PKGMAP)$(ADDITIONAL_SWG_FLAGS):. examples/proto/examplepb/*.proto

examples:


.PHONY: all
all: gateway swagger emp_protoc examples
