// Demo gRPC client: subscribes to task updates for a user and prints events.
// Usage: go run ./cmd/client -addr=localhost:50051 -user_id=1 -token=<JWT>
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/heth/STM/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	userID := flag.String("user_id", "1", "user ID to subscribe as (must match JWT)")
	token := flag.String("token", "", "Bearer JWT (required)")
	flag.Parse()

	if *token == "" {
		log.Fatal(" -token is required (JWT from /api/v1/auth/login)")
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := proto.NewNotificationServiceClient(conn)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	md := metadata.Pairs("authorization", "Bearer "+*token)
	ctx = metadata.NewOutgoingContext(ctx, md)

	stream, err := client.SubscribeToTaskUpdates(ctx, &proto.SubscribeRequest{UserId: *userID})
	if err != nil {
		log.Fatalf("SubscribeToTaskUpdates: %v", err)
	}

	log.Printf("Subscribed to task updates for user_id=%s (Ctrl+C to stop)\n", *userID)
	for {
		ev, err := stream.Recv()
		if err != nil {
			log.Printf("Recv error: %v", err)
			return
		}
		fmt.Printf("Event: %s | Task: id=%s title=%q status=%s\n",
			ev.GetType(), ev.GetTask().GetId(), ev.GetTask().GetTitle(), ev.GetTask().GetStatus())
	}
}
