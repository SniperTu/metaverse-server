package app

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	server "metaverse/app/servers"
	"metaverse/conf"
	"metaverse/logger"
	"metaverse/models"
	"metaverse/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func unaryEndToEndInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("%v", err)
		}
	}()
	md, _ := metadata.FromIncomingContext(ctx)
	var msgID string
	if msgIDs := md.Get("msgid"); msgIDs == nil || len(msgIDs) == 0 {
		msgID = utils.NewUUIDV4()
		md.Set("msgid", msgID)
		ctx = metadata.NewIncomingContext(ctx, md)
	} else {
		msgID = msgIDs[0]
	}
	bs, err := json.Marshal(req)
	if err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof(`[INNER REQ] msgid:%s, method:%s, metadata:%v, req:%s`, msgID, info.FullMethod, md, bs)
	}
	omd, _ := metadata.FromIncomingContext(ctx)
	m, err := handler(ctx, req)
	if err != nil {
		logger.Errorf("[ERROR RESP] msgid:%s, metadata:%v, error:%v", msgID, omd, err)
	} else {
		t := reflect.TypeOf(m)
		v := reflect.ValueOf(m)
		//如果m为指针类型并且未初始化导致返回nil错误，在此初始化
		if t.Kind() == reflect.Pointer && v.IsNil() {
			m = reflect.New(t.Elem()).Interface()
		}
		bs, _ := json.Marshal(m)
		if len(bs) != 0 {
			logger.Infof("[RESP] msgid:%s, metadata:%v, resp:%s", msgID, omd, bs)
		}
	}
	return m, err
}

// 用户校验拦截器
func unaryTokenCheckInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if server.SkipTokenCheck(ctx) {
		return handler(ctx, req)
	}
	switch info.FullMethod {
	case "/pbs.UserService/Login",
		"/pbs.UserService/AdminLogin",
		"/pbs.UserService/GetCode",
		"/pbs.UserService/SendCode",
		"/pbs.UserService/Register": //跳过token检查
	default:
		var token string
		token, err = server.TokenCheck(ctx)
		if err != nil {
			return
		}
		_, er := conf.RedisClient.Expire(models.RedisUserTokenCacheKeyPrefix+token, models.RedisUserTokenTTLSecs*time.Second).Result()
		if er != nil {
			logger.Errorf("msgid:%s,fresh token ttl failed!error:%v", server.MsgID(ctx), er)
		}
	}
	return handler(ctx, req)
}
