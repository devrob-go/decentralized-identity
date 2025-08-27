package main

import (
	"context"
	"log"
	"time"

	"api/auth/v1/proto"
	zlog "packages/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Initialize logger
	logger := zlog.NewLogger(zlog.Config{
		Level:      "debug",
		Output:     nil, // Use default stdout
		JSONFormat: false,
		AddCaller:  true,
		TimeFormat: time.RFC3339,
	})

	// Create context with correlation ID
	ctx := zlog.WithCorrelationID(context.Background(), "")

	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := proto.NewAuthServiceClient(conn)

	// Add correlation ID to metadata
	md := metadata.Pairs("x-correlation-id", "client-example-123")
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Example: Sign up a new user
	logger.Info(ctx, "Attempting to sign up user")
	signUpResp, err := client.SignUp(ctx, &proto.UserCreateRequest{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	})
	if err != nil {
		logger.Error(ctx, err, "SignUp failed", 400)
		log.Printf("SignUp failed: %v", err)
	} else {
		logger.Info(ctx, "SignUp successful", map[string]any{
			"user_id": signUpResp.User.Id,
			"email":   signUpResp.User.Email,
		})
		log.Printf("User created: %s (%s)", signUpResp.User.Name, signUpResp.User.Email)
	}

	// Example: Sign in
	logger.Info(ctx, "Attempting to sign in user")
	signInResp, err := client.SignIn(ctx, &proto.Credentials{
		Email:    "john.doe@example.com",
		Password: "password123",
	})
	if err != nil {
		logger.Error(ctx, err, "SignIn failed", 401)
		log.Printf("SignIn failed: %v", err)
	} else {
		logger.Info(ctx, "SignIn successful", map[string]any{
			"user_id": signInResp.User.Id,
		})
		log.Printf("User signed in: %s", signInResp.User.Name)
		log.Printf("Access token: %s", signInResp.Tokens.AccessToken[:20]+"...")
	}

	// Example: List users
	logger.Info(ctx, "Attempting to list users")
	listResp, err := client.ListUsers(ctx, &proto.ListUsersRequest{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		logger.Error(ctx, err, "ListUsers failed", 500)
		log.Printf("ListUsers failed: %v", err)
	} else {
		logger.Info(ctx, "ListUsers successful", map[string]any{
			"count": len(listResp.Users),
			"total": listResp.Total,
		})
		log.Printf("Found %d users (total: %d)", len(listResp.Users), listResp.Total)
		for _, user := range listResp.Users {
			log.Printf("  - %s (%s)", user.Name, user.Email)
		}
	}

	// Example: Refresh token (if we have tokens from sign in)
	if signInResp != nil && signInResp.Tokens != nil {
		logger.Info(ctx, "Attempting to refresh token")
		refreshResp, err := client.RefreshToken(ctx, &proto.RefreshTokenRequest{
			RefreshToken: signInResp.Tokens.RefreshToken,
		})
		if err != nil {
			logger.Error(ctx, err, "RefreshToken failed", 400)
			log.Printf("RefreshToken failed: %v", err)
		} else {
			logger.Info(ctx, "RefreshToken successful")
			log.Printf("New access token: %s", refreshResp.Tokens.AccessToken[:20]+"...")
		}
	}

	logger.Info(ctx, "Client example completed")
}
