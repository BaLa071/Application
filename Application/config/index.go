package config

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/sftp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/ssh"
)

func ConnectDB(ctx context.Context) (*mongo.Collection, *mongo.Collection) {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	Collection := client.Database("testing").Collection("Application")
	ChannelCollection := client.Database("testing").Collection("Channel")
	return Collection, ChannelCollection
}

func Sftp_connection() *sftp.Client {
	username := ""
	password := ""
	clientConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", "18.216.98.58:22", clientConfig)
	if err != nil {
		log.Fatalf("SSH DAIL FAILED:%v", err)
	}
	sftpClient, err1 := sftp.NewClient(conn)
	if err != nil {
		log.Fatalf("SFTP NEW CLIENT FAILED:%v\n", err1)
	} else {
		fmt.Println("done")
	}

	return sftpClient
}
