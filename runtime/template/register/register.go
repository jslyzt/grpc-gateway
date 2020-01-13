package register

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/jslyzt/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"net/http"
)

// HDStreaming 处理 stream
func HDStreaming(ctx context.Context, mux *runtime.ServeMux) func(http.ResponseWriter, *http.Request, runtime.Params) {
	return func(w http.ResponseWriter, req *http.Request, params runtime.Params) {
		err := status.Error(codes.Unimplemented, "streaming calls are not yet supported in the in-process transport")
		_, outbound := runtime.MarshalerForRequest(mux, req)
		runtime.HTTPError(ctx, mux, outbound, w, req, err)
	}
}

// HDForward 处理 Forward
func HDForward(ctx context.Context, mux *runtime.ServeMux,
	reqfn func(context.Context, runtime.Marshaler, *http.Request, runtime.Params) (proto.Message, runtime.ServerMetadata, error),
	respfn func(proto.Message) proto.Message) func(http.ResponseWriter, *http.Request, runtime.Params) {
	return func(w http.ResponseWriter, req *http.Request, params runtime.Params) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := reqfn(rctx, inboundMarshaler, req, params)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		if respfn != nil {
			resp = respfn(resp)
		}

		runtime.ForwardResponseMessage(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)
	}
}

// HDStreamForward 处理 stream Forward
func HDStreamForward(ctx context.Context, mux *runtime.ServeMux,
	reqfn func(context.Context, runtime.Marshaler, *http.Request, runtime.Params) (grpc.ClientStream, runtime.ServerMetadata, error),
	respfn func() proto.Message) func(http.ResponseWriter, *http.Request, runtime.Params) {
	return func(w http.ResponseWriter, req *http.Request, params runtime.Params) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := reqfn(rctx, inboundMarshaler, req, params)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		runtime.ForwardResponseStream(ctx, mux, outboundMarshaler, w, req, func() (proto.Message, error) {
			msg := respfn()
			err := resp.RecvMsg(msg)
			return msg, err
		}, mux.GetForwardResponseOptions()...)
	}
}

// RegFromEndpoint 注册
func RegFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption,
	deal func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error) error {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()
	return deal(ctx, mux, conn)
}
