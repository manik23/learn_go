package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	pb "learn-grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type contextKey string

const (
	ServerAddr                   = "localhost:50051"
	ServerTimeout                = 5 * time.Second
	ServerVersion                = "1.0.0"
	RequestAPI                   = "super-secret-key"
	RequestAPIKey     contextKey = "x-api-key"
	RequestVersionKey contextKey = "x-client-version"
	RequestIDKey      contextKey = "x-request-id"
	MetricsPort                  = ":2112"
)

type server struct {
	pb.UnimplementedGreeterServer
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func VersionInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovered: %v", r)
		}
	}()

	if ctx.Err() != nil {
		return nil, status.Errorf(codes.DeadlineExceeded, "deadline exceeded: %v", ctx.Err())
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is missing")
	}

	if err := validateAPIKey(md); err != nil {
		return nil, err
	}

	if err := validateVersion(md); err != nil {
		return nil, err
	}

	ctx = AddIDToCtx(ctx)

	return handler(ctx, req)
}

func VersionStreamInterceptor(
	srv any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovered: %v", r)
		}
	}()

	ctx := stream.Context()
	if ctx.Err() != nil {
		return status.Errorf(codes.DeadlineExceeded, "deadline exceeded: %v", ctx.Err())
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata is missing")
	}

	if err := validateAPIKey(md); err != nil {
		return err
	}

	if err := validateVersion(md); err != nil {
		return err
	}

	ctx = AddIDToCtx(ctx)

	stream = &wrappedStream{
		ServerStream: stream,
		ctx:          ctx,
	}

	return handler(srv, stream)
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	logRequestID(ctx)

	// Increment custom metric
	incrementTotalGreetings(ctx)

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
	logRequestID(stream.Context())

	// Increment custom metric for each chat message
	incrementTotalGreetings(stream.Context())

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

func (s *server) Chat(stream pb.Greeter_ChatServer) error {
	count := 0
	go func() {
		for {
			select {
			case <-stream.Context().Done():
				{
					log.Println("Chat Server Sending Error: ", stream.Context().Err().Error())
					return
				}

			case <-time.After(500 * time.Millisecond):
				{
					count++
					if err := stream.Send(&pb.HelloReply{
						Message:   "From Chat Server " + strconv.Itoa(count),
						Timestamp: timestamppb.Now(),
					}); err != nil {
						log.Println("Chat Server Sending Error: ", err.Error())
						return
					}
				}
			}
		}
	}()

	for {
		logRequestID(stream.Context())

		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Server Chat Received EOF: ", err.Error())
				return nil
			}

			if status.Code(err) == codes.Canceled {
				log.Println("Server Chat Received Canceled: ", err.Error())
				return status.Error(codes.Canceled, "client cancelled")
			}

			if status.Code(err) == codes.DeadlineExceeded {
				return status.Error(codes.DeadlineExceeded, "deadline exceeded")
			}

			log.Println("Server Chat Received Error: ", err.Error())
			return err
		}

		// Increment custom metric for each chat message
		incrementTotalGreetings(stream.Context())
		log.Printf("Chat Received: %v", req.GetName())
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			VersionInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
			VersionStreamInterceptor,
		),
	)

	// Register your gRPC service
	pb.RegisterGreeterServer(s, &server{})

	// Initialize all metrics
	grpc_prometheus.Register(s)

	// Register custom metrics
	registerCustomMetrics()

	// Register prometheus metrics handler
	http.Handle("/metrics", promhttp.Handler())

	// Start an HTTP server to expose metrics
	go func() {
		log.Printf("Metrics server listening at %s/metrics", MetricsPort)
		if err := http.ListenAndServe(MetricsPort, nil); err != nil {
			log.Fatalf("failed to serve metrics: %v", err)
		}
	}()

	// Start gRPC server and block until error
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
