package utilities

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JsonErr structures a standard error to return
type JsonErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

// JWTError is a struct that is used to contain a json encoded error message for any JWT related errors
type JWTError struct {
	Message string `json:"message"`
}

// GenerateObjectID for index keying records of data
func GenerateObjectID() string {
	newId := primitive.NewObjectID()
	return newId.Hex()
}

// CheckObjectID checks whether a hexID is null or now
func CheckObjectID(hexID string) bool {
	if hexID == "" || hexID == "000000000000000000000000" {
		return false
	}
	return true
}

// StrToBool converts a string to a bool
func StrToBool(in string) bool {
	if in == "true" || in == "TRUE" {
		return true
	}
	return false
}

// BoolToStr converts a bool to a string value
func BoolToStr(in bool) string {
	if in {
		return "true"
	}
	return "false"
}

// GetTokenFromContext parses an auth token from content metadata
func GetTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}
	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}
	return values[0], nil
}

// AttachTokenToContext inputs ctx and an auth token and returns ctx with token attached
func AttachTokenToContext(ctx context.Context, authToken string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", authToken)
}
