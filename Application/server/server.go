package main

import (
	"Application/config"
	services "Application/service"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "Application/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	services.Collection, services.ChannelCollection = config.ConnectDB(ctx)

	lis, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	defer cancel()
	s := grpc.NewServer()
	reflection.Register(s)
	fmt.Println("Listening")

	pb.RegisterApplicationServiceServer(s, &services.Server1{})
	pb.RegisterChannelServiceServer(s, &services.Server2{})
	pb.RegisterFileTransferServiceServer(s, &services.Server3{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
}
