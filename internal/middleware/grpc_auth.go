package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/heth/STM/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GrpcUserIDKey is the context key for the authenticated user ID (string) in gRPC handlers.
const GrpcUserIDKey = "grpc_user_id"

// JWTValidator validates access tokens and returns claims.
type JWTValidator interface {
	ValidateAccessToken(tokenString string) (*utils.JWTClaims, error)
}

// AuthUnaryInterceptor returns a unary interceptor that validates JWT from "authorization" metadata.
func AuthUnaryInterceptor(validator JWTValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userID, err := authFromContext(ctx, validator)
		if err != nil {
			return nil, err
		}
		return handler(context.WithValue(ctx, GrpcUserIDKey, userID), req)
	}
}

// AuthStreamInterceptor returns a stream interceptor that validates JWT from "authorization" metadata.
func AuthStreamInterceptor(validator JWTValidator) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userID, err := authFromContext(ss.Context(), validator)
		if err != nil {
			return err
		}
		wrapped := &streamWithContext{ServerStream: ss, ctx: context.WithValue(ss.Context(), GrpcUserIDKey, userID)}
		return handler(srv, wrapped)
	}
}

type streamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *streamWithContext) Context() context.Context {
	return s.ctx
}

// authFromContext reads Authorization metadata, validates the Bearer token, and returns the user ID as string.
func authFromContext(ctx context.Context, validator JWTValidator) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization required")
	}
	parts := strings.SplitN(strings.TrimSpace(vals[0]), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header")
	}
	claims, err := validator.ValidateAccessToken(strings.TrimSpace(parts[1]))
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid or expired token")
	}
	return strings.TrimSpace(claimUserID(claims)), nil
}

func claimUserID(c *utils.JWTClaims) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%d", c.UserID)
}
