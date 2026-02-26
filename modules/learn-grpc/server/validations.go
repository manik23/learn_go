package main

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func validateAPIKey(md metadata.MD) error {
	apiKeys := md.Get(string(RequestAPIKey))

	// Add API Key check
	if len(apiKeys) == 0 {
		return status.Error(codes.Unauthenticated, "api key is missing")
	}

	if apiKeys[0] != RequestAPI {
		return status.Errorf(codes.Unauthenticated, "invalid api key: %s", apiKeys[0])
	}
	return nil
}

func validateVersion(md metadata.MD) error {
	versions := md.Get(string(RequestVersionKey))
	if len(versions) == 0 {
		return status.Error(codes.InvalidArgument, "client version is missing")
	}

	if versions[0] != ServerVersion {
		return status.Errorf(codes.InvalidArgument, "invalid client version: %s", versions[0])
	}
	return nil
}
