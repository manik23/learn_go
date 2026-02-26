package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func logRequestID(ctx context.Context) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}

	requestIDs := md.Get(string(RequestIDKey))
	if len(requestIDs) >= 1 {
		log.Printf("Processing request for Gopher in handler [RequestID: %s]", requestIDs[0])
	} else {
		log.Println("missing request id")
	}
}

func AddIDToCtx(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	requestIDs := md.Get(string(RequestIDKey))
	if len(requestIDs) == 0 || requestIDs[0] == "" {
		md.Set(string(RequestIDKey), uuid.New().String())
	}

	return metadata.NewIncomingContext(ctx, md)
}
