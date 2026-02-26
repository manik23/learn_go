package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "learn-grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ClientAddr    = "localhost:50051"
	ClientTimeout = 5 * time.Second
	ClientVersion = "1.0.0"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(ClientAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Unary RPC
	log.Printf("Calling SayHello...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "Gopher", Version: &pb.Version{Version: ClientVersion}})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	// Server Streaming RPC
	log.Printf("Calling StreamHello...")
	stream, err := c.StreamHello(context.Background(), &pb.HelloRequest{Name: "Gopher", Version: &pb.Version{Version: ClientVersion}})
	if err != nil {
		log.Fatalf("could not open stream: %v", err)
	}
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.StreamHello(_) = _, %v", c, err)
		}
		log.Printf("Stream Reply: %s", reply.GetMessage())
	}
}
