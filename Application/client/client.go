package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "Application/proto"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewFileTransferServiceClient(conn)
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	ctx := context.TODO()
	err1 := Get(ctx, client)
	if err1 != nil {
		log.Fatal("error client service: ", err1)
	}
}

func Get(ctx context.Context, client pb.FileTransferServiceClient) error {
	GetFileReq := &pb.GetRequest{
		AppId:  "653f4e61960643a90d0ad2f5",
		FoldId: "708a3b6e-0a6e-48f4-aa00-4316e2542109",
	}
	GetFileRes, err := client.Get(ctx, GetFileReq)
	if err != nil {
		log.Println(err)
	}
	LocalFile, _ := os.Create("/home/balaji/go/src/newtest.json")
	count := 1
	for {
		req, err := GetFileRes.Recv()
		if err == io.EOF {
			fmt.Println("EOF!!")
			break
		}
		if err != nil {
			return err
		}
		chunk := req.GetChunk()
		_, err1 := LocalFile.Write(chunk)
		if err != nil {
			log.Println("error in writing ", err1)
		}
		fmt.Println(count)
		count++
	}
	fmt.Println(GetFileRes)
	return nil
}

func Put(ctx context.Context, client pb.FileTransferServiceClient) error {
	stream, err := client.Put(ctx)
	if err != nil {
		log.Println(err)
	}
	LocalFile, _ := os.Open("/home/balaji/go/src/test.json")
	stat, _ := LocalFile.Stat()
	buffer := make([]byte, stat.Size())
	count := 1
	for {
		num, err := LocalFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("error reading file: ", err)
		}
		chunk := buffer[:num]
		if err := stream.Send(&pb.PutRequest{
			AppId:  "653f4e61960643a90d0ad2f5",
			FoldId: "708a3b6e-0a6e-48f4-aa00-4316e2542109",
			Chunk:  chunk}); err != nil {
			log.Println(err)
		}
		fmt.Println(count)
		count++
		time.Sleep(5 * time.Second)
	}
	// time.Sleep(15 * time.Second)
	return nil
}
