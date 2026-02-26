package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "learn-grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	ServerAddr    = "localhost:50051"
	ServerTimeout = 5 * time.Second
	ServerVersion = "1.0.0"
)

type contextKey string

type server struct {
	pb.UnimplementedGreeterServer
}

func VersionInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if ctx.Err() != nil {
		return nil, status.Errorf(codes.DeadlineExceeded, "deadline exceeded: %v", ctx.Err())
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is missing")
	}

	versions := md.Get("x-client-version")
	if len(versions) == 0 {
		return nil, status.Error(codes.InvalidArgument, "client version is missing")
	}

	if versions[0] != ServerVersion {
		return nil, status.Errorf(codes.InvalidArgument, "invalid client version: %s", versions[0])
	}

	log.Printf("Unary Interceptor [%s]: Client Version %s", info.FullMethod, versions[0])
	return handler(ctx, req)
}

func VersionStreamInterceptor(
	srv any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := stream.Context()
	if ctx.Err() != nil {
		return status.Errorf(codes.DeadlineExceeded, "deadline exceeded: %v", ctx.Err())
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata is missing")
	}

	versions := md.Get("x-client-version")
	if len(versions) == 0 {
		return status.Error(codes.InvalidArgument, "client version is missing")
	}

	if versions[0] != ServerVersion {
		return status.Errorf(codes.InvalidArgument, "invalid client version: %s", versions[0])
	}

	log.Printf("Stream Interceptor [%s]: Client Version %s", info.FullMethod, versions[0])
	return handler(srv, stream)
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, status.Errorf(codes.DeadlineExceeded, "deadline exceeded: %v", ctx.Err())
	case <-timer.C:
		return &pb.HelloReply{
			Message:   "Hello " + in.GetName(),
			Timestamp: timestamppb.Now(),
		}, nil
	}
}

func (s *server) StreamHello(in *pb.HelloRequest, stream pb.Greeter_StreamHelloServer) error {
	log.Printf("Streaming to: %v", in.GetName())
	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("Hello %s (message %d)", in.GetName(), i+1)
		if stream.Context().Err() != nil {
			return stream.Context().Err()
		}
		if err := stream.Send(&pb.HelloReply{
			Message:   msg,
			Timestamp: timestamppb.Now(),
		}); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(VersionInterceptor),
		grpc.StreamInterceptor(VersionStreamInterceptor),
	)
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
