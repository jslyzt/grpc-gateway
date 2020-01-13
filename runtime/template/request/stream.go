package request

import (
	"github.com/jslyzt/grpc-gateway/runtime"
	"github.com/jslyzt/grpc-gateway/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
)

// StreamSend 发送
func StreamSend(stream grpc.ClientStream, preq interface{}, marshaler runtime.Marshaler, req *http.Request) (mdata *runtime.ServerMetadata, err error) {
	dec := marshaler.NewDecoder(req.Body)
	for {
		err = dec.Decode(preq)
		if err == io.EOF {
			break
		}
		if err != nil {
			grpclog.Infof("Failed to decode request: %v", err)
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		if err = stream.SendMsg(preq); err != nil {
			if err == io.EOF {
				break
			}
			grpclog.Infof("Failed to send request: %v", err)
			return nil, err
		}
	}

	if err := stream.CloseSend(); err != nil {
		grpclog.Infof("Failed to terminate client stream: %v", err)
		return nil, err
	}
	mdata, err = StreamGainMetadata(stream)
	return
}

// StreamGainMetadata 获取Metadata
func StreamGainMetadata(stream grpc.ClientStream) (*runtime.ServerMetadata, error) {
	header, err := stream.Header()
	if err != nil {
		grpclog.Infof("Failed to get header from client: %v", err)
		return nil, err
	}
	mdata := &runtime.ServerMetadata{
		HeaderMD: header,
	}
	return mdata, nil
}

// StreamDecode 解析
func StreamDecode(marshaler runtime.Marshaler, req *http.Request, preq interface{}) (func() io.Reader, error) {
	reader, err := utilities.IOReaderFactory(req.Body)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err = marshaler.NewDecoder(reader()).Decode(preq); err != nil && err != io.EOF {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return reader, nil
}

// StreamSend 发送
func StreamBiDiSend(stream grpc.ClientStream, preq interface{}, marshaler runtime.Marshaler, req *http.Request) (mdata *runtime.ServerMetadata, err error) {
	dec := marshaler.NewDecoder(req.Body)
	handleSend := func() error {
		err := dec.Decode(preq)
		if err == io.EOF {
			return err
		}
		if err != nil {
			grpclog.Infof("Failed to decode request: %v", err)
			return err
		}
		if err := stream.SendMsg(preq); err != nil {
			grpclog.Infof("Failed to send request: %v", err)
			return err
		}
		return nil
	}
	if err := handleSend(); err != nil {
		if cerr := stream.CloseSend(); cerr != nil {
			grpclog.Infof("Failed to terminate client stream: %v", cerr)
		}
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	go func() {
		for {
			if err := handleSend(); err != nil {
				break
			}
		}
		if err := stream.CloseSend(); err != nil {
			grpclog.Infof("Failed to terminate client stream: %v", err)
		}
	}()
	mdata, err = StreamGainMetadata(stream)
	return mdata, nil
}