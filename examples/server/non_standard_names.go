package server

import (
	"context"

	"github.com/golang/glog"
	examples "github.com/jslyzt/grpc-gateway/examples/proto/examplepb"
)

// Implements NonStandardServiceServer

type nonStandardServer struct{}

func newNonStandardServer() examples.NonStandardServiceServer {
	return new(nonStandardServer)
}

// Update 更新
func (s *nonStandardServer) Update(ctx context.Context, msg *examples.NonStandardUpdateRequest) (*examples.NonStandardMessage, error) {
	glog.Info(msg)

	newMsg := &examples.NonStandardMessage{
		// The fieldmask_helper doesn't generate nested structs if they are nil
		Thing: &examples.NonStandardMessage_Thing{SubThing: &examples.NonStandardMessage_Thing_SubThing{}},
	}
	applyFieldMask(newMsg, msg.Body, msg.UpdateMask)

	glog.Info(newMsg)
	return newMsg, nil
}

// UpdateWithJSONNames 更新
func (s *nonStandardServer) UpdateWithJSONNames(ctx context.Context, msg *examples.NonStandardWithJSONNamesUpdateRequest) (*examples.NonStandardMessageWithJSONNames, error) {
	glog.Info(msg)

	newMsg := &examples.NonStandardMessageWithJSONNames{
		// The fieldmask_helper doesn't generate nested structs if they are nil
		Thing: &examples.NonStandardMessageWithJSONNames_Thing{SubThing: &examples.NonStandardMessageWithJSONNames_Thing_SubThing{}},
	}
	applyFieldMask(newMsg, msg.Body, msg.UpdateMask)

	glog.Info(newMsg)
	return newMsg, nil
}
