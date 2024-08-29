package models

import (
	"context"
	"fmt"
	"metaverse/app/model"
	"metaverse/conf"
	"metaverse/logger"
	"metaverse/pbs"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserLoginSummary struct {
	pbs.UserLoginSummary
	model.Model
}

var UserLoginSummaryColl *mongo.Collection

func init() {
	UserLoginSummaryColl = conf.Mongo.Collection("user_login_summary")
}

// 获取登录汇总信息
func (uls *UserLoginSummary) GetBySchoolID(schoolId string) (rs *pbs.UserLoginSummary, err error) {
	var ret *mongo.SingleResult
	rs = new(pbs.UserLoginSummary)
	if len(schoolId) == 0 {
		// 不属于任何学校
		ret = UserLoginSummaryColl.FindOne(context.Background(), bson.M{
			"school_id": bson.M{
				"$exists": false,
			},
		})
	} else {
		ret = UserLoginSummaryColl.FindOne(context.Background(), bson.D{
			{"school_id", schoolId},
		})
	}
	if err = ret.Err(); err != nil {
		logger.Errorf("%v", err)
		if err == mongo.ErrNoDocuments {
			return rs, nil
		}
		return
	}
	if err = ret.Decode(rs); err != nil {
		logger.Errorf("%v", err)
		return nil, err
	}
	return
}

// 用户登录汇总信息参数自增操作
func (uls *UserLoginSummary) Increase(loginUserCount, loginCount, loginMins int, schoolId string) (after *pbs.UserLoginSummary, err error) {
	after = new(pbs.UserLoginSummary)
	update := bson.M{
		"$inc": bson.M{
			"login_user_total_count": loginUserCount,
			"login_total_count":      loginCount,
			"login_total_duration":   loginMins,
		},
		"$setOnInsert": bson.M{"created_at": time.Now().Unix()},
		"$set":         bson.M{"updated_at": time.Now().Unix()},
	}
	filter := bson.M{}
	if len(schoolId) == 0 {
		filter["school_id"] = bson.M{"$exists": false}
	} else {
		filter["school_id"] = schoolId
	}
	err = UserLoginSummaryColl.FindOneAndUpdate(context.Background(), filter, update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(after)
	if err != nil {
		logger.Errorf("%v", err)
	}
	return
}

func init() {
	// 初始化登录用户总数
	matchStage := bson.D{
		{"$match", bson.D{
			{"last_login", bson.D{{"$gt", 0}}},
		}},
	}
	countStage := bson.D{
		{"$count", "loginCount"},
	}
	// 初始化累计登录人数
	cursor, err := UserColl.Aggregate(context.Background(), []bson.D{matchStage, countStage})
	if err != nil {
		fmt.Printf("UserLoginSummary init failed! err: %v\n", err)
		return
	}
	lc := []bson.M{}
	if err = cursor.All(context.Background(), &lc); err != nil {
		fmt.Printf("UserLoginSummary init failed! err: %v\n", err)
		return
	}
	var loginCount int32
	if len(lc) > 0 {
		loginCount, _ = lc[0]["loginCount"].(int32)
	}
	if _, err = UserLoginSummaryColl.UpdateOne(context.Background(), bson.M{}, bson.M{
		"$set": bson.M{"login_user_total_count": loginCount},
	}, options.Update().SetUpsert(true)); err != nil {
		fmt.Printf("UserLoginSummary init failed! err:%v\n", err)
	}
}
