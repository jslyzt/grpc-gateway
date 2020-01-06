package server

import (
	"context"
	"io"

	examples "github.com/jslyzt/grpc-gateway/examples/proto/examplepb"
)

type flowCombinationServer struct{}

func newFlowCombinationServer() examples.FlowCombinationServer {
	return &flowCombinationServer{}
}

// RpcEmptyRpc 空rpc
func (s flowCombinationServer) RpcEmptyRpc(ctx context.Context, req *examples.EmptyProto) (*examples.EmptyProto, error) {
	return req, nil
}

// RpcEmptyStream 空stream
func (s flowCombinationServer) RpcEmptyStream(req *examples.EmptyProto, stream examples.FlowCombination_RpcEmptyStreamServer) error {
	return stream.Send(req)
}

// StreamEmptyRpc stream空rpc
func (s flowCombinationServer) StreamEmptyRpc(stream examples.FlowCombination_StreamEmptyRpcServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return stream.SendAndClose(new(examples.EmptyProto))
}

// StreamEmptyStream stream空stream
func (s flowCombinationServer) StreamEmptyStream(stream examples.FlowCombination_StreamEmptyStreamServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return stream.Send(new(examples.EmptyProto))
}

// RpcBodyRpc body rpc
func (s flowCombinationServer) RpcBodyRpc(ctx context.Context, req *examples.NonEmptyProto) (*examples.EmptyProto, error) {
	return new(examples.EmptyProto), nil
}

// RpcPathSingleNestedRpc path single nested rpc
func (s flowCombinationServer) RpcPathSingleNestedRpc(ctx context.Context, req *examples.SingleNestedProto) (*examples.EmptyProto, error) {
	return new(examples.EmptyProto), nil
}

// RpcPathNestedRpc path nested rpc
func (s flowCombinationServer) RpcPathNestedRpc(ctx context.Context, req *examples.NestedProto) (*examples.EmptyProto, error) {
	return new(examples.EmptyProto), nil
}

// RpcBodyStream rpc body stream
func (s flowCombinationServer) RpcBodyStream(req *examples.NonEmptyProto, stream examples.FlowCombination_RpcBodyStreamServer) error {
	return stream.Send(new(examples.EmptyProto))
}

// RpcPathSingleNestedStream rpc path single nested stream
func (s flowCombinationServer) RpcPathSingleNestedStream(req *examples.SingleNestedProto, stream examples.FlowCombination_RpcPathSingleNestedStreamServer) error {
	return stream.Send(new(examples.EmptyProto))
}

// RpcPathNestedStream rpc path nested stream
func (s flowCombinationServer) RpcPathNestedStream(req *examples.NestedProto, stream examples.FlowCombination_RpcPathNestedStreamServer) error {
	return stream.Send(new(examples.EmptyProto))
}
