package server

import (
	"context"

	examples "github.com/jslyzt/grpc-gateway/examples/proto/examplepb"
)

// Implements of ResponseBodyServiceServer

type responseBodyServer struct{}

func newResponseBodyServer() examples.ResponseBodyServiceServer {
	return new(responseBodyServer)
}

// GetResponseBody 获取
func (s *responseBodyServer) GetResponseBody(ctx context.Context, req *examples.ResponseBodyIn) (*examples.ResponseBodyOut, error) {
	return &examples.ResponseBodyOut{
		Response: &examples.ResponseBodyOut_Response{
			Data: req.Data,
		},
	}, nil
}

// ListResponseBodies 列表
func (s *responseBodyServer) ListResponseBodies(ctx context.Context, req *examples.ResponseBodyIn) (*examples.RepeatedResponseBodyOut, error) {
	return &examples.RepeatedResponseBodyOut{
		Response: []*examples.RepeatedResponseBodyOut_Response{
			&examples.RepeatedResponseBodyOut_Response{
				Data: req.Data,
			},
		},
	}, nil
}

// ListResponseStrings 列表
func (s *responseBodyServer) ListResponseStrings(ctx context.Context, req *examples.ResponseBodyIn) (*examples.RepeatedResponseStrings, error) {
	if req.Data == "empty" {
		return &examples.RepeatedResponseStrings{
			Values: []string{},
		}, nil
	}
	return &examples.RepeatedResponseStrings{
		Values: []string{"hello", req.Data},
	}, nil
}
