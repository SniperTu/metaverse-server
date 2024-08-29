package server

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"metaverse/conf"
	"metaverse/models"
	"metaverse/pbs"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/go-redis/redis"
	"github.com/mojocn/base64Captcha"
	gCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
}

// 获取token
func (this Server) Token(ctx context.Context) (str string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	token, ok := md["token"]
	if !ok {
		return
	}
	str = token[0]
	return
}

// 获取session
func (this Server) Sess(ctx context.Context) (str string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	sess, ok := md["session"]
	if !ok {
		return
	}
	str = sess[0]
	return
}

func MsgID(ctx context.Context) (str string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	exMsgId, ok := md["msgid"]
	if len(exMsgId) == 0 || !ok {
		return
	}
	return exMsgId[0]
}

// 元数据中skiptoken有 jumpjumpjump 标识时跳过token验证,用于非登录请求但要跳过token验证
func SkipTokenCheck(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok && len(md["skiptoken"]) != 0 && md["skiptoken"][0] == "jumpjumpjump" {
		return true
	}
	return false
}

// 用户认证
func (this Server) Auth(ctx context.Context) (user models.User, err error) {
	var token = this.Token(ctx)
	if token == "" {
		err = status.Error(gCodes.Unauthenticated, "未登录")
		return
	}
	return models.GetTokenCache(token)
}

func TokenCheck(ctx context.Context) (token string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = status.Error(gCodes.Unauthenticated, "登录令牌缺失")
		return
	}
	exToken := md.Get("token")
	if len(exToken) == 0 || len(exToken[0]) == 0 {
		err = status.Error(gCodes.Unauthenticated, "登录令牌缺失")
		return
	}
	token = exToken[0]
	_, err = models.GetTokenCache(token)
	return
}

// 存储用户认证
func (this Server) StoreTokenLocalCache(token string, user pbs.User) {
	models.SetTokenCache(token, models.User{User: user})
}

// 移除用户认证
func (this Server) RemoveTokenLocalCache(ctx context.Context) {
	var token = this.Token(ctx)
	if token == "" {
		return
	}
	models.RemoveTokenCache(token)
}

// 支持多种验证码格式
type configJsonBody struct {
	Id            string
	CaptchaType   string
	VerifyValue   string
	DriverAudio   *base64Captcha.DriverAudio
	DriverString  *base64Captcha.DriverString
	DriverChinese *base64Captcha.DriverChinese
	DriverMath    *base64Captcha.DriverMath
	DriverDigit   *base64Captcha.DriverDigit
}

// 存储用户验证码
func (this Server) Code(ctx context.Context) (code, idKey string, err error) {
	var driver base64Captcha.Driver
	var param configJsonBody = configJsonBody{
		CaptchaType: "string",
		DriverString: &base64Captcha.DriverString{
			Length:          4,
			Height:          50,
			Width:           100,
			ShowLineOptions: 2,
			NoiseCount:      0,
			Source:          "1234567890abcdefghijklmnopqrstuvwxyz",
		},
	}
	switch param.CaptchaType {
	case "audio":
		driver = param.DriverAudio
	case "string":
		driver = param.DriverString.ConvertFonts()
	case "math":
		driver = param.DriverMath.ConvertFonts()
	case "chinese":
		driver = param.DriverChinese.ConvertFonts()
	default:
		driver = param.DriverDigit
	}
	store := base64Captcha.DefaultMemStore
	c := base64Captcha.NewCaptcha(driver, store)
	// 获取
	idKey, code, _, err = c.Generate()
	return
}

// 检测用户验证码
func (this Server) CheckCode(ctx context.Context, code string) bool {
	sess := this.Sess(ctx)
	if sess == "" {
		return false
	}
	return base64Captcha.DefaultMemStore.Verify(sess, code, true)
}

func (this Server) SendSmsCode(mobile string, captchaType int32) (err error) {
	t := time.Now().UnixNano()
	rand.New(rand.NewSource(t + 1))
	code1 := rand.Intn(10)
	rand.New(rand.NewSource(t + 2))
	code2 := rand.Intn(10)
	rand.New(rand.NewSource(t + 3))
	code3 := rand.Intn(10)
	rand.New(rand.NewSource(t + 4))
	code4 := rand.Intn(10)
	code := fmt.Sprintf("%d%d%d%d", code1, code2, code3, code4)
	client, err := dysmsapi.NewClientWithAccessKey("cn-Shanghai", "", "")
	if err != nil {
		return
	}
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"

	request.PhoneNumbers = mobile
	request.SignName = ""
	var smsTemplateCode string
	if captchaType == 0 { //注册发送验证码
		smsTemplateCode = "SMS_193140680"
	} else if captchaType == 1 { //登录发送验证码
		smsTemplateCode = "SMS_193235189"
	}
	request.TemplateCode = smsTemplateCode
	request.TemplateParam = fmt.Sprintf("{\"code\":\"%s\"}", code)
	response, err := client.SendSms(request)
	if err != nil {
		return
	}
	if response.Code != "OK" {
		err = errors.New("验证码发送失败")
		return
	}
	conf.RedisClient.Set(fmt.Sprintf("code_%s", mobile), code, 120*time.Second)
	return
}

func (this Server) CheckSmsCode(mobile, code string) error {
	val, err := conf.RedisClient.Get(fmt.Sprintf("code_%s", mobile)).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if val == "" {
		return errors.New("验证码已过期，请重新获取")
	}
	if val != code {
		return errors.New("验证码不正确")
	}
	conf.RedisClient.Del(fmt.Sprintf("code_%s", mobile))
	return nil
}
