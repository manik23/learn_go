package main

import (
	"context"
	"errors"
	"io"
	"log"
	"math/rand"
	"time"

	pb "learn-grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	ClientAddr                   = "localhost:50051"
	ClientTimeout                = 5 * time.Second
	ClientVersion                = "1.0.0"
	RequestAPI                   = "super-secret-key"
	RequestAPIKey     contextKey = "x-api-key"
	RequestVersionKey contextKey = "x-client-version"
	RequestIDKey      contextKey = "x-request-id"
)

func setupMetadata(ctx context.Context) context.Context {
	ctx = metadata.AppendToOutgoingContext(ctx, string(RequestVersionKey), ClientVersion)
	ctx = metadata.AppendToOutgoingContext(ctx, string(RequestAPIKey), RequestAPI)
	return ctx
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.NewClient(ClientAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Unary RPC
	log.Printf("Calling SayHello...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rand.Intn(3))*time.Second)
	defer cancel()
	ctx = setupMetadata(ctx)

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "Gopher"})
	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			log.Printf("deadline exceeded during SayHello: %s", err.Error())
		} else {
			log.Fatalf("could not greet: %s", err.Error())
		}
	}
	log.Printf("Greeting: %s", r.GetMessage())

	// Server Streaming RPC
	log.Printf("Calling StreamHello...")
	streamCtx, streamCancel := context.WithTimeout(
		context.Background(),
		time.Duration(rand.Intn(7))*time.Second,
	)
	// Add Metadata
	streamCtx = setupMetadata(streamCtx)
	defer streamCancel()

	stream, err := c.StreamHello(
		streamCtx,
		&pb.HelloRequest{Name: "Gopher"},
	)
	if err != nil {
		log.Fatalf("could not open stream: %v", err)
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("stream closed")
			} else {
				switch status.Code(err) {
				case codes.DeadlineExceeded:
					log.Printf("deadline exceeded during StreamHello: %s", err.Error())
				case codes.InvalidArgument:
					log.Printf("invalid argument during StreamHello: %s", err.Error())
				case codes.Unimplemented:
					log.Printf("unimplemented during StreamHello: %s", err.Error())
				default:
					log.Fatalf("%v.StreamHello(_) = _, %v", c, err)
				}
			}
			break
		}

		log.Printf(
			"Stream Reply: %s : %s",
			reply.GetMessage(),
			reply.GetTimestamp().AsTime().Format(time.RFC1123),
		)
	}
}
