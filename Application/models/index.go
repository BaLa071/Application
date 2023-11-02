package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Application struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	ChannelId string             `json:"channelId" bson:"channelId"`
	CreatedAt string             `json:"createdAt" bson:"createdAt"`
	UpdatedAt string             `json:"updatedAt" bson:"updatedAt"`
	CreatedBy string             `json:"createdBy" bson:"createdBy"`
}

type AddChannelReq struct {
	ID                 primitive.ObjectID `bson:"_id"`
	Name               string             `bson:"name"`
	ChannelType        string             `bson:"channelType"`
	ServerIP           string             `bson:"serverIp"`
	AuthenticationType string             `bson:"authenticationType"`
	Username           string             `bson:"username,omitempty"`
	Password           string             `bson:"password,omitempty"`
	PrivateKey         string             `bson:"privateKey,omitempty"`
	Folders            []Fold             `bson:"fold"`
}

type Fold struct {
	FoldID   string `bson:"id"`
	Path     string `bson:"path"`
	GPGKeyID string `bson:"gpgkeyid"`
}

type ChanReq struct {
	ChannelID string `bson:"channelID"`
}
