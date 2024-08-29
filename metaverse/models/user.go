package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"metaverse/app/model"
	"metaverse/conf"
	"metaverse/pbs"
	"metaverse/utils"
	"reflect"
	"time"

	"metaverse/logger"

	"github.com/go-redis/redis"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	//初始化自动检查超级管理员账号并自动创建参数常量
	DEFAULT_SUPER_ADMIN_USERNAME  = "admin"       //默认超级管理员用户名
	DEFAULT_SUPER_ADMIN_PASSWORD  = "123456"      //默认超级管理员密码
	DEFAULT_SUPER_ADMIN_MOBILE    = "13122738739" //默认超级管理员手机号
	DEFAULT_SUPER_ADMIN_USERFNAME = "超级管理员"       //默认超级管理员姓名
)

const (
	RedisUserTokenCacheKeyPrefix = "MetaversionUserToken:" //用户token信息redis缓存key前缀(string类型)
	RedisUserTokenTTLSecs        = 15 * 60                 //用户token信息redis缓存过期时间(秒)
)

type User struct {
	pbs.User
	model.Model
}

var UserColl *mongo.Collection //集合

func (U *User) Create(data *pbs.User) error {
	data.Id = primitive.NewObjectID().Hex()
	data.CreatedAt = time.Now().Unix()
	return U.SetColl(UserColl).Add(data)
}
func (U *User) Delete(data *pbs.User) (err error) {
	data.DeletedAt = time.Now().Unix()
	return U.Update(data)
}

func (U *User) Update(data *pbs.User) error {
	data.UpdatedAt = time.Now().Unix()
	update := utils.Struct2Map(*data)
	delete(update, "_id")
	delete(update, "created_at")
	delete(update, "_id,omitempty")
	_, err := UserColl.UpdateOne(context.Background(), bson.M{"_id": data.Id}, bson.M{"$set": update})
	if err != nil {
		logger.Errorf("%v", err)
	}
	return err
}

func (U *User) GetUserById(userId string) (rs *pbs.User, err error) {
	rs = new(pbs.User)
	err = UserColl.FindOne(context.Background(), bson.M{"_id": userId, "deleted_at": 0}).Decode(&rs)
	if err != nil {
		logger.Errorf("%v", err)
	}
	return
}

func (U *User) GetUserByUserName(userName string) (rs *pbs.User, err error) {
	rs = new(pbs.User)
	var filter = bson.M{"deleted_at": 0}
	filter["or"] = []bson.M{{"user_name": userName}, {"mobile": userName}}
	err = UserColl.FindOne(context.Background(), filter).Decode(&rs)
	if err != nil {
		logger.Errorf("%v", err)
	}
	return
}

func (U *User) GetUserByMobile(mobile string) (rs *pbs.User, err error) {
	rs = new(pbs.User)
	err = UserColl.FindOne(context.Background(), bson.M{"mobile": mobile, "deleted_at": 0}).Decode(&rs)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Errorf("%v", err)
		} else {
			err = nil
		}
	}
	return
}

func (U *User) List(name, mobile, schoolName string, pageIdx, pageSize int64, userType pbs.UserType, mustSchool, mustLogined, hideImg bool) (rs []*pbs.User, totalPages, totalRows int64, err error) {
	filter := bson.M{
		"deleted_at": 0,
	}
	filter["user_type"] = bson.D{{"$lte", userType}}
	if len(name) != 0 {
		filter["user_fname"] = bson.D{{"$regex", name}}
	}
	if len(mobile) != 0 {
		filter["mobile"] = bson.D{{"$regex", mobile}}
	}
	if mustLogined {
		filter["last_login"] = bson.D{{"$gt", 0}}
	}
	if len(schoolName) != 0 {
		if mustSchool {
			filter["school_info.name"] = schoolName
		} else {
			filter["$or"] = bson.A{
				bson.M{
					"school_info.name": schoolName,
				},
				bson.M{
					"school_info.name": bson.M{
						"$exists": false,
					},
				},
				bson.M{
					"school_info.name": "",
				},
			}
		}
	}
	var matchStage bson.D
	if len(filter) != 0 {
		matchStage = bson.D{
			{"$match", filter},
		}
	}
	countStage := bson.D{
		{"$count", "count"},
	}
	sortStage := bson.D{
		{"$sort", bson.D{{"last_login", -1}}},
	}
	if pageSize == 0 {
		pageSize = 10
	}
	if pageIdx <= 1 {
		pageIdx = 1
	}
	pageStage := []bson.D{
		{
			{"$skip", (pageIdx - 1) * pageSize},
		},
		{
			{"$limit", pageSize},
		},
	}
	//若为-1，不进行分页
	if pageSize == -1 {
		pageStage = []bson.D{}
	}
	var cCursor *mongo.Cursor
	stages := []bson.D{}
	if matchStage != nil && len(matchStage) != 0 {
		stages = append(stages, matchStage)
	}
	cCursor, err = UserColl.Aggregate(context.Background(), append(stages, countStage))
	if err != nil {
		logger.Errorf("%v", err)
		return
	}
	crs := []struct {
		Count int32
	}{}
	if err = cCursor.All(context.Background(), &crs); err != nil {
		logger.Errorf("%v", err)
		return
	}
	if len(crs) != 0 && crs[0].Count > 0 {
		totalRows = int64(crs[0].Count)
	}
	totalPages = totalRows/pageSize + (totalRows%pageSize+pageSize-1)/pageSize
	//过滤头像文件
	if hideImg {
		projectStage := bson.D{
			{"$project", bson.D{{"roles.avatar", 0}}},
		}
		stages = append(stages, projectStage)
	}
	var mCursor *mongo.Cursor
	mCursor, err = UserColl.Aggregate(context.Background(), append(append(stages, sortStage), pageStage...))
	if err != nil {
		logger.Errorf("%v", err)
		return
	}
	if pageSize > totalRows || pageSize == -1 {
		pageSize = totalRows
	}
	rs = make([]*pbs.User, pageSize)
	for k := range rs {
		rs[k] = new(pbs.User)
	}
	if err = mCursor.All(context.Background(), &rs); err != nil {
		logger.Errorf("%v", err)
		return
	}
	return
}

func init() {
	UserColl = conf.Mongo.Collection("user")
	admin := pbs.User{}
	if err := UserColl.FindOne(context.Background(), bson.M{"user_name": "admin"}).Decode(&admin); err != nil {
		if err != mongo.ErrNoDocuments {
			fmt.Println("super admin user init() finding failed! ", err)
			return
		}
		// 自动创建
		admin.Mobile = DEFAULT_SUPER_ADMIN_MOBILE
		admin.UserFname = DEFAULT_SUPER_ADMIN_USERFNAME
		admin.Password = utils.Password(DEFAULT_SUPER_ADMIN_PASSWORD)
		admin.UserName = DEFAULT_SUPER_ADMIN_USERNAME
		admin.UserType = pbs.UserType_superAdmin
		if err = new(User).Create(&admin); err != nil {
			fmt.Println("super admin user creation failed! ", err)
		}
		return
	}
}

func GetTokenCache(token string) (u User, err error) {
	if len(token) == 0 {
		err = status.Error(codes.Unauthenticated, "登录令牌缺失")
		return
	}
	var ustr string
	ustr, err = conf.RedisClient.Get(RedisUserTokenCacheKeyPrefix + token).Result()
	if err != nil {
		logger.Errorf("%v", err)
		if err == redis.Nil {
			err = status.Error(codes.Unauthenticated, "登录过期或者无效")
		}
		return
	}
	if len(ustr) == 0 {
		err = status.Error(codes.Unauthenticated, "登录过期或者无效")
		return
	}
	if err = json.Unmarshal([]byte(ustr), &u.User); err != nil {
		logger.Errorf("%v", err)
		err = status.Error(codes.Unauthenticated, "登录令牌解析失败")
	}
	return
}

func SetTokenCache(token string, u User) (err error) {
	if len(token) == 0 {
		logger.Errorf("SetTokenCache failed! empty token, userId:%s", u.Id)
		return
	}
	eu := pbs.User{}
	if reflect.DeepEqual(u.User, eu) {
		logger.Errorf("user info empty,save cache failed!")
		return
	}
	bs := []byte{}
	bs, err = json.Marshal(&u.User)
	if err != nil {
		logger.Errorf("%v", err)
		return
	}
	_, err = conf.RedisClient.Set(RedisUserTokenCacheKeyPrefix+token, string(bs), RedisUserTokenTTLSecs*time.Second).Result()
	if err != nil {
		logger.Errorf("%v", err)
		err = errors.New("user token cache set failed!")
		return
	}
	return
}

func RemoveTokenCache(token string) (err error) {
	if len(token) == 0 {
		err = errors.New("token missing")
		return
	}
	_, err = conf.RedisClient.Del(RedisUserTokenCacheKeyPrefix + token).Result()
	if err != nil {
		logger.Errorf("%v", err)
		err = errors.New("remove token cahine failed")
	}
	return
}

func FreshTokenFromDBWithToken(token string) (err error) {
	var u User
	u, err = GetTokenCache(token)
	if err != nil {
		return
	}
	var newUser *pbs.User
	newUser, err = u.GetUserById(u.Id)
	if err != nil {
		return
	}
	u.User = *newUser
	return SetTokenCache(token, u)
}
