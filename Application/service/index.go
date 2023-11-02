package services

import (
	"Application/config"
	"Application/models"
	pb "Application/proto"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server1 struct {
	pb.UnimplementedApplicationServiceServer
}

type Server2 struct {
	pb.UnimplementedChannelServiceServer
}

type Server3 struct {
	pb.UnimplementedFileTransferServiceServer
}

var Collection, ChannelCollection *mongo.Collection

func (s *Server1) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	time := time.Now()
	application := &models.Application{
		Id:        primitive.NewObjectID(),
		Name:      req.Name,
		ChannelId: req.ChannelId,
		CreatedAt: time.Format("2006-01-02 15:04:05"),
		UpdatedAt: time.Format("2006-01-02 15:04:05"),
		CreatedBy: req.CreatedBy,
	}
	res, err := Collection.InsertOne(ctx, application)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	fmt.Println(res)
	return &pb.CreateResponse{ChannelId: req.ChannelId}, nil
}

func (s *Server1) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	id, _ := primitive.ObjectIDFromHex(req.Id)
	filter := bson.M{"_id": id}
	var application *models.Application
	res, err := Collection.Find(ctx, filter)
	if err != nil {
		log.Println("err in finding: ", err)
		return nil, err
	}
	for res.Next(ctx) {
		if err := res.Decode(&application); err != nil {
			log.Println("err in decoding: ", err)
			return nil, err
		}
	}
	stringObjectID := application.Id.Hex()
	response := &pb.ListResponse{
		Id:        stringObjectID,
		Name:      application.Name,
		ChannelId: application.ChannelId,
		CreatedBy: application.CreatedBy,
		CreatedAt: application.CreatedAt,
		UpdatedAt: application.UpdatedAt,
	}
	return response, nil
}

var CurrentPages = 1

func (s *Server1) ListAll(ctx context.Context, in *pb.Empty) (*pb.ListAllResponse, error) {
	limit := 10
	opts := options.Count().SetHint(bson.M{})
	count, err := Collection.CountDocuments(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	temp := int(count) % int(limit)
	pages := (int(count) / int(limit))
	if temp != 0 {
		pages += 1
	}
	if pages < CurrentPages {
		CurrentPages = 1
	}
	skip := (CurrentPages - 1) * limit
	fmt.Println(skip)
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(skip))

	res, err := Collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var applications []*pb.ListResponse
	for res.Next(ctx) {
		var app models.Application
		if err := res.Decode(&app); err != nil {
			log.Println("err", err)
		}
		appID := app.Id.Hex()
		temp := &pb.ListResponse{
			Id:        appID,
			Name:      app.Name,
			ChannelId: app.ChannelId,
			CreatedBy: app.CreatedBy,
			CreatedAt: app.CreatedAt,
			UpdatedAt: app.UpdatedAt,
		}
		applications = append(applications, temp)
	}
	if err := res.Err(); err != nil {
		log.Fatal(err)
	}
	res.Close(ctx)
	CurrentPages++
	response := &pb.ListAllResponse{
		ListAll: applications,
	}
	return response, nil
}

func (s *Server2) AddChannel(ctx context.Context, req *pb.AddChannelReq) (*pb.ViewChannelRes, error) {

	var folders []models.Fold
	for _, i := range req.Folders {
		fold := models.Fold{
			FoldID:   uuid.New().String(),
			Path:     i.Path,
			GPGKeyID: i.GPGKeyID,
		}
		folders = append(folders, fold)
	}

	db := &models.AddChannelReq{
		ID:                 primitive.NewObjectID(),
		Name:               req.Name,
		ChannelType:        req.ChannelType.String(),
		ServerIP:           req.ServerIP,
		AuthenticationType: req.AuthenticationType,
		Username:           req.Username,
		Password:           req.Password,
		PrivateKey:         req.PrivateKey,
		Folders:            folders,
	}

	//adding a channel
	_, err := ChannelCollection.InsertOne(ctx, db)
	if err != nil {
		fmt.Println("can't insert channel:  ", err)
		return nil, err
	}

	var chanInfo models.AddChannelReq
	query := bson.M{"_id": db.ID}
	find := ChannelCollection.FindOne(ctx, query)

	temp := find.Decode(&chanInfo)
	if temp != nil {
		fmt.Println("can't find channels:  ", err)
		return nil, err
	}

	var foldersResp []*pb.FoldResp
	for _, i := range chanInfo.Folders {
		foldResp := &pb.FoldResp{
			FoldID:   i.FoldID,
			Path:     i.Path,
			GPGKeyID: i.GPGKeyID,
		}
		foldersResp = append(foldersResp, foldResp)
	}

	res := &pb.ViewChannelRes{
		ID:                 chanInfo.ID.Hex(),
		Name:               chanInfo.Name,
		ChannelType:        chanInfo.ChannelType,
		ServerIP:           chanInfo.ServerIP,
		AuthenticationType: chanInfo.AuthenticationType,
		Username:           chanInfo.Username,
		Password:           chanInfo.Password,
		PrivateKey:         chanInfo.PrivateKey,
		FoldersResp:        foldersResp,
	}

	return res, nil
}

func (s *Server2) ViewChannel(ctx context.Context, req *pb.ChanReq) (*pb.ViewChannelRes, error) {

	objId, err := primitive.ObjectIDFromHex(req.ChannelID)
	if err != nil {
		fmt.Println("error in converting objID: ", err)
		return nil, err
	}

	var chanInfo models.AddChannelReq
	query := bson.M{"_id": objId}
	find := ChannelCollection.FindOne(ctx, query).Decode(&chanInfo)
	if find == mongo.ErrNoDocuments {
		fmt.Println("no doc exits: ", find)
		return nil, err
	}
	if find != nil {
		fmt.Println("error in decoding chanInfo: ", find)
		return nil, err
	}

	id := req.ChannelID
	var foldersResp []*pb.FoldResp
	for _, i := range chanInfo.Folders {
		fold := &pb.FoldResp{
			FoldID:   i.FoldID,
			Path:     i.Path,
			GPGKeyID: i.GPGKeyID,
		}
		foldersResp = append(foldersResp, fold)
	}

	res := &pb.ViewChannelRes{
		ID:                 id,
		Name:               chanInfo.Name,
		ChannelType:        chanInfo.ChannelType,
		ServerIP:           chanInfo.ServerIP,
		AuthenticationType: chanInfo.AuthenticationType,
		Username:           chanInfo.Username,
		Password:           chanInfo.Password,
		PrivateKey:         chanInfo.PrivateKey,
		FoldersResp:        foldersResp,
	}

	return res, nil
}

// for pagenation
var CurrentPage = 1

func (s *Server2) ListChannel(ctx context.Context, emp *pb.EmptyReq) (*pb.ListChannelRes, error) {

	count, err := ChannelCollection.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	//3 records per page
	limit := int64(3)
	skip := int64(CurrentPage-1) * limit
	fOpt := options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}

	temp := count % limit
	pages := 0
	if temp != 0 {
		pages = (int(count) / int(limit)) + 1
	} else {
		pages = (int(count) / int(limit))
	}

	if pages < CurrentPage {
		CurrentPage = 1
	}

	fmt.Println(count, " ", CurrentPage, " ", skip)
	CurrentPage++

	find, err := ChannelCollection.Find(ctx, bson.M{}, &fOpt)
	if err != nil {
		fmt.Println("error in finding all docs: ", err)
		return nil, err
	}

	var chanList []*pb.ViewChannelRes
	var chanInfo models.AddChannelReq
	for find.Next(ctx) {
		err := find.Decode(&chanInfo)
		if err != nil {
			fmt.Println("error in decoding to chaninfo: ", err)
			return nil, err
		}

		id := chanInfo.ID.Hex()
		var foldersResp []*pb.FoldResp
		for _, i := range chanInfo.Folders {
			fold := &pb.FoldResp{
				FoldID:   i.FoldID,
				Path:     i.Path,
				GPGKeyID: i.GPGKeyID,
			}
			foldersResp = append(foldersResp, fold)
		}

		chanl := &pb.ViewChannelRes{
			ID:                 id,
			Name:               chanInfo.Name,
			ChannelType:        chanInfo.ChannelType,
			ServerIP:           chanInfo.ServerIP,
			AuthenticationType: chanInfo.AuthenticationType,
			Username:           chanInfo.Username,
			Password:           chanInfo.Password,
			PrivateKey:         chanInfo.PrivateKey,
			FoldersResp:        foldersResp,
		}

		chanList = append(chanList, chanl)
	}
	res := &pb.ListChannelRes{
		ListResp: chanList,
	}

	return res, nil
}

func (s *Server3) Put(stream pb.FileTransferService_PutServer) error {
	var ctx context.Context
	SftpClient := config.Sftp_connection()
	temp := 0
	var err error
	var remoteFile *sftp.File
	for {
		res, err1 := stream.Recv()
		if err1 == io.EOF {
			break
		}
		if err1 != nil {
			// fmt.Println("error pls check")
			log.Println("err: ", err1)
			break
		}
		if temp == 0 {
			remoteFile, err = Finding(ctx, SftpClient, res)
			if err != nil {
				log.Println("err in fetching data!!")
				return err
			}
		}
		chunk := res.GetChunk()
		_, err2 := remoteFile.Write(chunk)
		if err2 != nil {
			log.Fatal("error writing chunk: ", err2)
		}
		temp++
		fmt.Println(temp)
	}
	fmt.Println(temp)
	return stream.SendAndClose(&pb.PutResponse{Messag: "Done!!"})
}

func Finding(ctx context.Context, sftpClient *sftp.Client, req *pb.PutRequest) (*sftp.File, error) {
	var app models.Application
	AppID, _ := primitive.ObjectIDFromHex(req.AppId)
	filter := bson.M{"_id": AppID}
	err1 := Collection.FindOne(ctx, filter).Decode(&app)
	if err1 != nil {
		return nil, err1
	}
	var channel models.AddChannelReq
	ChannelID, _ := primitive.ObjectIDFromHex(app.ChannelId)
	filter = bson.M{"_id": ChannelID}
	err := ChannelCollection.FindOne(ctx, filter).Decode(&channel)
	if err != nil {
		return nil, err
	}
	var FoldPath string
	for _, i := range channel.Folders {
		if i.FoldID == req.FoldId {
			FoldPath = i.Path
		}
	}
	if FoldPath == " " {
		log.Println("Path is empty!")
	}
	remoteFile, err := sftpClient.Create(FoldPath) 
	if err != nil {
		return nil, err
	}
	fmt.Println("remotefile")
	return remoteFile, nil
}

func (s *Server3) Get(req *pb.GetRequest, stream pb.FileTransferService_GetServer) error {
	var ctx context.Context
	SftpClient := config.Sftp_connection()

	var app models.Application
	AppID, _ := primitive.ObjectIDFromHex(req.AppId)
	filter := bson.M{"_id": AppID}
	err1 := Collection.FindOne(ctx, filter).Decode(&app)
	if err1 != nil {
		return err1
	}
	var channel models.AddChannelReq
	ChannelID, _ := primitive.ObjectIDFromHex(app.ChannelId)
	filter = bson.M{"_id": ChannelID}
	err := ChannelCollection.FindOne(ctx, filter).Decode(&channel)
	if err != nil {
		return err
	}
	var FoldPath string
	for _, i := range channel.Folders {
		if i.FoldID == req.FoldId {
			FoldPath = i.Path
		}
	}
	if FoldPath == " " {
		log.Println("Path is empty!")
	}
	remoteFile, err := SftpClient.Open(FoldPath)
	if err != nil {
		return err
	}
	// stat, _ := remoteFile.Stat()
	buffer := make([]byte, 50)
	for {
		num, err := remoteFile.Read(buffer)
		if err == io.EOF {
			chunk := buffer[:num]
			fmt.Println(chunk)
			if err := stream.Send(&pb.GetResponse{Chunk: chunk, Messag: "done!"}); err != nil {
				log.Fatal("error sending file: ", err)
			}
			break
		}
		if err != nil {
			log.Fatal("error reading file: ", err)
		}
		chunk := buffer[:num]
		fmt.Println(chunk)
		if err := stream.Send(&pb.GetResponse{Chunk: chunk, Messag: "done!"}); err != nil {
			log.Fatal("error sending file: ", err)
		}
	}
	return nil
}

// {
//     "AppId": "653f4e61960643a90d0ad2f5",
//     "FoldId": "708a3b6e-0a6e-48f4-aa00-4316e2542109"
// }
