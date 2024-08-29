package models

import (
	"context"
	"metaverse/app/model"
	"metaverse/conf"
	"metaverse/pbs"
	"metaverse/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	pbs.Config
	model.Model
}

var ConfigColl *mongo.Collection

func init() {
	ConfigColl = conf.Mongo.Collection("config")
}

func (C *Config) Update(data *pbs.Config) error {
	update := utils.Struct2Map(data)
	_, err := ConfigColl.UpdateOne(context.Background(), bson.M{"_id": data.Id}, bson.M{"$set": update})
	return err
}
