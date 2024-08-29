package conf

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var Mongo = newMongo()

func newMongo() *mongo.Database {
	var err error
	opt := options.Client()
	opt.Hosts = Conf.Db.Mongo.Hosts
	if Conf.Db.Mongo.Pwd != "" {
		opt.Auth = &options.Credential{
			AuthMechanism: "SCRAM-SHA-1",
			AuthSource:    Conf.Db.Mongo.Database,
			Username:      Conf.Db.Mongo.User,
			Password:      Conf.Db.Mongo.Pwd,
			PasswordSet:   true,
		}
	}
	opt.SetLocalThreshold(time.Second * 3). //只使用与mongo操作耗时小于3秒的
						SetMaxConnIdleTime(5 * time.Millisecond). //指定连接可以保持空闲的最大毫秒数
						SetMaxPoolSize(200)                       //使用最大的连接数
	opt.ReadConcern = readconcern.Majority()
	opt.WriteConcern = writeconcern.Majority()
	var client *mongo.Client
	ctx, cancel := getContext()
	defer cancel()
	if client, err = mongo.Connect(ctx, opt); err != nil {
		checkErr(err)
	}
	return client.Database(Conf.Db.Mongo.Database)
}

func checkErr(err error) {
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("没有查到数据")
			os.Exit(0)
		} else {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

func getContext() (ctx context.Context, cancel context.CancelFunc) {
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	return
}
