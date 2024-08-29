package servers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	server "metaverse/app/servers"
	"metaverse/conf"
	"metaverse/logger"
	"metaverse/models"
	"metaverse/pbs"
	"metaverse/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	server.Server
}

// 用户登录接口封装,登录接口统一调用，提供userType合法校验
func (U *UserServer) loginWithValidType(ctx context.Context, req *pbs.LoginUser, incrLoginTime bool, validUserType ...pbs.UserType) (resp *pbs.User, err error) {
	if req.UserName == "" {
		err = errors.New("请输入登录用户名")
		return
	}
	if req.Password == "" {
		err = errors.New("请输入密码")
		return
	}
	if req.Code == "" {
		err = errors.New("请输入验证码")
		return
	}
	if U.CheckCode(ctx, req.Code) {
		err = errors.New("验证码不正确")
		return
	}
	userInfo, err := new(models.User).GetUserByUserName(req.UserName)
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	if userInfo == nil {
		err = errors.New("该用户不存在")
		return
	}
	if userInfo.Password != utils.Password(req.Password) {
		err = errors.New("密码错误")
		return
	}
	if validUserType != nil && len(validUserType) != 0 {
		var isValid bool
		for _, v := range validUserType {
			if v == userInfo.UserType {
				isValid = true
			}
		}
		if !isValid {
			err = status.Error(codes.InvalidArgument, "用户无权登录")
			return
		}
	}
	var token = utils.Md5(fmt.Sprintf("%dyuanyuzhou", time.Now().UnixNano()))
	userInfo.Token = token
	if incrLoginTime {
		orinLastLogin := userInfo.LastLogin
		nowSec := time.Now().Unix()
		userInfo.LastLogin = nowSec
		userInfo.LoginTimes += 1
		var ldm int64
		// 有过登录记录的情况才可能记录登录时长
		if orinLastLogin != 0 {
			// 上次登录未退出，以两次登录时间作为登录时长增量
			if userInfo.LastLogoutTime <= orinLastLogin {
				ldm = (nowSec - orinLastLogin) / 60
				ldm++
				userInfo.LoginDurationMin += ldm
			}
		}
		var loginUserCount int
		// 未登录过增加累计登录人数
		if orinLastLogin == 0 {
			loginUserCount = 1
		}
		// 更新全局汇总信息
		if _, er := new(models.UserLoginSummary).Increase(loginUserCount, 1, int(ldm), ""); er != nil {
			logger.Errorf("%v", er)
		}
		// 更新学校汇总信息
		if userInfo.SchoolInfo != nil && len(userInfo.SchoolInfo.Id) != 0 && userInfo.UserType <= pbs.UserType_admin {
			if _, er := new(models.UserLoginSummary).Increase(loginUserCount, 1, int(ldm), userInfo.SchoolInfo.Id); er != nil {
				logger.Errorf("%v", er)
			}
		}
		new(models.User).Update(userInfo)
	}
	U.StoreTokenLocalCache(token, *userInfo)
	resp = userInfo
	return
}

// 管理后台登录
func (U *UserServer) AdminLogin(ctx context.Context, req *pbs.LoginUser) (resp *pbs.User, err error) {
	return U.loginWithValidType(ctx, req, false, pbs.UserType_admin, pbs.UserType_superAdmin)
}

// 登录接口
func (U *UserServer) Login(ctx context.Context, req *pbs.LoginUser) (resp *pbs.User, err error) {
	return U.loginWithValidType(ctx, req, true)
}

// 用户登出，更新登录时长和登录汇总信息
func (U *UserServer) Logout(ctx context.Context, req *pbs.LogoutReq) (resp *pbs.Empty, err error) {
	msgId := server.MsgID(ctx)
	var loginMinsIncr int64 = 1 //登录时长增量初始化为1
	var user *pbs.User

	nowSec := time.Now().Unix()
	if len(req.UserId) != 0 {
		user, err = new(models.User).GetUserById(req.UserId)
		if err != nil {
			err = status.Error(codes.Internal, "get user info failed! msgid: "+msgId)
			return
		}
		// 若已经从其他端登出，则通过最后登出时间开始算
		if user.LastLogoutTime > user.LastLogin {
			loginMinsIncr += (nowSec - user.LastLogoutTime) / 60
		} else {
			loginMinsIncr += (nowSec - user.LastLogin) / 60
		}
		user.LastLogoutTime = nowSec
		user.LoginDurationMin += loginMinsIncr
		if err = new(models.User).Update(user); err != nil {
			err = status.Error(codes.Internal, "user login time info update failed!msgid: "+msgId)
			return
		}
		//更新登录汇总信息表,全局汇总数据更新
		if _, err = new(models.UserLoginSummary).Increase(0, 0, int(loginMinsIncr), ""); err != nil {
			err = status.Error(codes.Internal, "user login summary update failed!msgid: "+msgId)
		}
		//更新登录汇总信息表,学校汇总数据更新，超级管理员不更新
		if user.SchoolInfo != nil && len(user.SchoolInfo.Id) != 0 && user.UserType <= pbs.UserType_admin {
			if _, err = new(models.UserLoginSummary).Increase(0, 0, int(loginMinsIncr), user.SchoolInfo.Id); err != nil {
				err = status.Error(codes.Internal, "user login school summary update failed!msgid: "+msgId)
			}
		}
	} else {
		// 登出，不更新登录时长和汇总时长，清理token缓存
		var mu models.User
		mu, err = U.Auth(ctx)
		if err != nil {
			return
		}
		token := mu.Token
		mu.Token = "null"
		if err = mu.Update(&mu.User); err != nil {
			err = status.Error(codes.Internal, "user login time info update failed!msgid: "+msgId)
			return
		}
		if err = models.RemoveTokenCache(token); err != nil {
			err = status.Error(codes.Internal, "user token cache clean failed!msgid: "+msgId)
			return
		}
	}
	// 刷新更新后的用户信息的token
	if err = models.FreshTokenFromDBWithToken(user.Id); err != nil {
		err = status.Error(codes.Internal, "fresh user info cache failed!msgid: "+msgId)
		return
	}
	return
}

func (U *UserServer) GetCode(ctx context.Context, _ *pbs.Empty) (resp *pbs.CodeResp, err error) {
	resp = new(pbs.CodeResp)
	code, idKey, err := U.Code(ctx)
	if err != nil {
		return
	}
	resp.Code = code
	resp.Id = idKey
	return
}

// 创建用户接口，学校管理员只能创建同学校的账户
func (U *UserServer) Create(ctx context.Context, req *pbs.CreateUserReq) (resp *pbs.Empty, err error) {
	loginUser, err := U.Auth(ctx)
	if err != nil {
		return
	}
	if len(req.UserName) == 0 {
		err = status.Error(codes.InvalidArgument, "用户名不能为空")
		return
	}
	if len(req.Mobile) == 0 {
		err = status.Error(codes.InvalidArgument, "手机号不能为空")
		return
	}
	if !utils.VerifyMobile(req.Mobile) {
		err = status.Error(codes.InvalidArgument, "请输入正确的手机号")
		return
	}
	if len(req.Password) == 0 {
		err = status.Error(codes.InvalidArgument, "请输入注册密码")
		return
	}
	if req.UserType == pbs.UserType_UserType_Omit {
		err = status.Error(codes.InvalidArgument, "用户类型缺失")
		return
	}
	var ui *pbs.User
	ui, err = new(models.User).GetUserByMobile(req.Mobile)
	if err != nil {
		err = status.Error(codes.Internal, "查询相同手机号用户信息失败")
		return
	}
	if ui != nil {
		err = status.Error(codes.InvalidArgument, "该手机号已被注册")
		return
	}
	ui = new(pbs.User)
	ui.UserName = req.UserName
	ui.UserFname = req.UserFname
	ui.Mobile = req.Mobile
	ui.Password = utils.Password(req.Password)
	ui.UserType = req.UserType
	ui.ClientConfig = &pbs.ClientConfig{}
	if loginUser.SchoolInfo != nil {
		ui.SchoolInfo = loginUser.SchoolInfo
	}
	err = new(models.User).Create(ui)
	resp = new(pbs.Empty)
	return
}

func (U *UserServer) Update(ctx context.Context, req *pbs.UpdateUserReq) (resp *pbs.Empty, err error) {
	msgId := server.MsgID(ctx)
	loginUser, err := U.Auth(ctx)
	if err != nil {
		return
	}
	if len(req.UserId) == 0 {
		err = status.Error(codes.InvalidArgument, "id missing")
		return
	}
	var ui *pbs.User
	ui, err = new(models.User).GetUserById(req.UserId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = status.Error(codes.NotFound, "用户未找到,Id:"+req.UserId)
			return
		}
		err = status.Error(codes.Internal, "查询用户信息失败")
		return
	}
	if len(req.UserName) != 0 {
		ui.UserName = req.UserName
	}
	if len(req.UserFname) != 0 {
		ui.UserFname = req.UserFname
	}
	if len(req.Mobile) != 0 {
		if !utils.VerifyMobile(req.Mobile) {
			err = status.Error(codes.InvalidArgument, "请输入正确的手机号")
			return
		}
		ui.Mobile = req.Mobile
	}
	if len(req.Password) != 0 {
		ui.Password = utils.Password(req.Password)
	}
	if loginUser.UserType == pbs.UserType_nomral || loginUser.UserType == pbs.UserType_admin {
		if req.UserType == pbs.UserType_superAdmin {
			err = status.Error(codes.InvalidArgument, "普通用户/学校管理员用户禁止提升超级管理员权限")
			return
		}
		if ui.UserType == pbs.UserType_superAdmin {
			err = status.Error(codes.InvalidArgument, "普通用户/学校管理员用户禁止修改超级管理员信息")
			return
		}
	}
	ui.UserType = req.UserType
	if len(req.Avatar) != 0 {
		ui.Avatar = req.Avatar
	}
	ui.Gender = req.Gender
	err = new(models.User).Update(ui)
	if mongo.IsDuplicateKeyError(err) {
		err = status.Error(codes.AlreadyExists, "用户手机号已存在")
		return
	}
	// 刷新更新后的用户信息token信息
	if err = models.FreshTokenFromDBWithToken(ui.Id); err != nil {
		err = status.Error(codes.Internal, "fresh user info cache failed!msgid: "+msgId)
		return
	}
	return
}

func (U *UserServer) Register(ctx context.Context, req *pbs.RegisterReq) (resp *pbs.Empty, err error) {
	if req.Mobile == "" {
		err = errors.New("手机号不能为空")
		return
	}
	if !utils.VerifyMobile(req.Mobile) {
		err = errors.New("请输入正确的手机号")
		return
	}
	if req.Password == "" {
		err = errors.New("请输入注册密码")
		return
	}
	if req.Code == "" {
		err = errors.New("验证码不能为空")
		return
	}
	if userInfos, errs := new(models.User).GetUserByMobile(req.Mobile); userInfos != nil && errs == nil {
		errs = errors.New("该手机号已被注册")
		return
	}
	err = U.CheckSmsCode(req.Mobile, req.Code)
	if err != nil {
		return
	}
	var userInfo = new(pbs.User)
	userInfo.Mobile = req.Mobile
	userInfo.Password = utils.Password(req.Password)
	// 客户端默认音量开关：开，音量大小：50
	userInfo.ClientConfig = &pbs.ClientConfig{
		Voice:  1,
		Volume: 50,
	}
	userInfo.UserType = pbs.UserType_nomral
	err = new(models.User).Create(userInfo)
	resp = new(pbs.Empty)
	return
}

func (U *UserServer) Delete(ctx context.Context, req *pbs.DeleteUserReq) (resp *pbs.Empty, err error) {
	msgId := server.MsgID(ctx)
	if len(req.UserId) == 0 {
		err = status.Error(codes.InvalidArgument, "id missing!msgid: "+msgId)
		return
	}
	ui, er := new(models.User).GetUserById(req.UserId)
	if er != nil {
		logger.Errorf("msgid:%s, err:%v", msgId, er)
		err = status.Error(codes.Internal, "用户删除失败，msgid:"+msgId)
		return
	}
	if err = new(models.User).Delete(ui); err != nil {
		logger.Errorf("msgid:%s, err:%v", msgId, err)
		err = status.Error(codes.Internal, "用户删除失败，msgid:"+msgId)
		return
	}
	return
}

func (U *UserServer) SendCode(ctx context.Context, req *pbs.SendCodeReq) (resp *pbs.Empty, err error) {
	resp = new(pbs.Empty)
	if req.Mobile == "" {
		err = errors.New("请输入手机号")
		return
	}
	if !utils.VerifyMobile(req.Mobile) {
		err = errors.New("请输入正确的手机号")
		return
	}
	userInfo, err := new(models.User).GetUserByMobile(req.Mobile)
	if userInfo != nil && err == nil {
		err = errors.New("该手机号已被注册")
		return
	}
	err = U.SendSmsCode(req.Mobile, 0)
	return
}

func (U *UserServer) GetLoginSummary(ctx context.Context, req *pbs.UserLoginSummaryReq) (resp *pbs.UserLoginSummaryResp, err error) {
	loginUser, _ := U.Auth(ctx)
	var schoolId string
	if loginUser.SchoolInfo != nil && loginUser.UserType <= pbs.UserType_admin {
		schoolId = loginUser.SchoolInfo.Id
	}
	resp = new(pbs.UserLoginSummaryResp)
	resp.PageNumber = req.PageNumber
	resp.Summary, err = new(models.UserLoginSummary).GetBySchoolID(schoolId)
	if err != nil {
		err = status.Error(codes.Internal, "获取登录汇总信息失败")
		return
	}
	var schoolName string
	if loginUser.SchoolInfo != nil && loginUser.UserType <= pbs.UserType_admin {
		schoolName = loginUser.SchoolInfo.Name
	}
	var users []*pbs.User
	users, resp.TotalPages, resp.TotalRows, err = new(models.User).List("", "", schoolName, req.PageNumber, req.PageSize, loginUser.UserType, true, true, true)
	if err != nil {
		err = status.Error(codes.Internal, "获取用户列表失败")
		return
	}
	resp.UserList = make([]*pbs.UserLoginSummaryRespUser, len(users))
	for i, ui := range users {
		resp.UserList[i] = new(pbs.UserLoginSummaryRespUser)
		var roleName string
		if len(ui.Roles) != 0 {
			rns := make([]string, len(ui.Roles))
			for i, r := range ui.Roles {
				rns[i] = r.Name
			}
			roleName = strings.Join(rns, ",")
		}
		resp.UserList[i].RoleName = roleName
		resp.UserList[i].Mobile = ui.Mobile
		resp.UserList[i].LastLogin = ui.LastLogin
		resp.UserList[i].LoginTimes = ui.LoginTimes
		resp.UserList[i].LoginDurationMin = ui.LoginDurationMin
	}
	return
}

func (U *UserServer) GetList(ctx context.Context, req *pbs.UserListReq) (resp *pbs.UserList, err error) {
	loginUser, _ := U.Auth(ctx)
	var schoolName string
	var mustSchool bool
	//如果登录账户包含学校信息且不是超级管理员则附加学校筛选
	if loginUser.SchoolInfo != nil && loginUser.UserType != pbs.UserType_superAdmin {
		schoolName = loginUser.SchoolInfo.Name
		mustSchool = true
	}
	//过滤搜索条件首尾空白
	req.UserFname = strings.Trim(req.UserFname, " ")
	req.Mobile = strings.Trim(req.Mobile, " ")
	schoolName = strings.Trim(schoolName, " ")
	// 逻辑处理
	users, tp, tr, err := new(models.User).List(req.UserFname, req.Mobile, schoolName, req.PageNumber, req.PageSize, loginUser.UserType, mustSchool, false, true)
	if err != nil {
		err = status.Error(codes.Internal, "获取用户列表失败")
		return
	}
	resp = &pbs.UserList{
		Users:      users,
		TotalPages: tp,
		TotalRows:  tr,
		PageNumber: req.PageNumber,
	}
	return
}

func (U *UserServer) GetMyAccountInfo(ctx context.Context, _ *pbs.Empty) (resp *pbs.User, err error) {
	var user models.User
	msgId := server.MsgID(ctx)
	user, err = U.Auth(ctx)
	if err != nil {
		return
	}
	resp, err = user.GetUserById(user.Id)
	if err != nil {
		logger.Errorf("msgId:%s,%v", msgId, err)
		err = status.Error(codes.NotFound, "用户获取失败")
	}
	return
}

// 用户禁言接口
func (U *UserServer) BanToPost(ctx context.Context, req *pbs.UserSwitchReq) (resp *pbs.Empty, err error) {
	msgID := server.MsgID(ctx)
	if len(req.UserId) == 0 {
		err = status.Error(codes.InvalidArgument, "id missing!msgid: "+msgID)
		return
	}
	_, err = models.UserColl.UpdateOne(context.Background(), bson.M{
		"_id": req.UserId,
	}, bson.M{"$set": bson.M{
		"banned_to_post": req.TurnOn,
		"updated_at":     time.Now().Unix(),
	}})
	if err != nil {
		logger.Errorf("msgid:%s,err:%v", msgID, err)
		err = status.Error(codes.Internal, "禁言用户失败！msgid:"+msgID)
	}
	var status string
	if req.TurnOn {
		status = "/1"
	} else {
		status = "/0"
	}
	var rsp *http.Response
	rsp, err = http.Get(strings.TrimRight(conf.Conf.InteractiveServerHTTPAddr, "/") + "/banNotify/" + req.UserId + "/0" + status)
	if err != nil {
		logger.Errorf("banToPost notify failed!error: %v", err)
		return
	}
	if rsp.StatusCode != http.StatusOK {
		bs := []byte{}
		if _, err = rsp.Body.Read(bs); err != nil {
			logger.Errorf("banToPost notify failed!body read error:%v", err)
			return
		}
		logger.Errorf("banToPost notify failed!status:%d,msg:%s", rsp.StatusCode, bs)
	}
	return
}

// 用户禁用接口
func (U *UserServer) Ban(ctx context.Context, req *pbs.UserSwitchReq) (resp *pbs.Empty, err error) {
	msgID := server.MsgID(ctx)
	if len(req.UserId) == 0 {
		err = status.Error(codes.InvalidArgument, "id missing!msgid: "+msgID)
		return
	}
	var ustatus int32
	if req.TurnOn {
		ustatus = 1 //禁用
	}
	_, err = models.UserColl.UpdateOne(context.Background(), bson.M{
		"_id": req.UserId,
	}, bson.M{"$set": bson.M{
		"status":     ustatus,
		"updated_at": time.Now().Unix(),
	}})
	if err != nil {
		logger.Errorf("msgid:%s,err:%v", msgID, err)
		err = status.Error(codes.Internal, "用户更新失败！msgid:"+msgID)
	}

	var rsp *http.Response
	reqURL := strings.TrimRight(conf.Conf.InteractiveServerHTTPAddr, "/") + "/banNotify/" + req.UserId + "/1/" + strconv.Itoa(int(ustatus))
	rsp, err = http.Get(reqURL)
	if err != nil {
		logger.Errorf("ban notify failed!error: %v", err)
		return
	}
	if rsp.StatusCode != http.StatusOK {
		bs := []byte{}
		if _, err = rsp.Body.Read(bs); err != nil {
			logger.Errorf("ban notify failed!body read error:%v", err)
			return
		}
		logger.Errorf("ban notify failed!status:%d,msg:%s", rsp.StatusCode, bs)
	}
	return
}

func (U *UserServer) ClientSysConfig(ctx context.Context, req *pbs.ClientConfigReq) (resp *pbs.Empty, err error) {
	msgId := server.MsgID(ctx)
	if len(req.UserId) == 0 {
		err = status.Error(codes.InvalidArgument, "user_id missing!msgid: "+msgId)
		return
	}
	if req.ClientConfig == nil {
		err = status.Error(codes.InvalidArgument, "client_config param missing!msgid: "+msgId)
		return
	}
	updateBM := bson.M{}
	bs, _ := json.Marshal(req.ClientConfig)
	json.Unmarshal(bs, &updateBM)
	_, err = models.UserColl.UpdateOne(ctx, bson.M{
		"_id": req.UserId,
	}, bson.M{"$set": bson.M{
		"client_config": updateBM,
		"updated_at":    time.Now().Unix(),
	}})
	if err != nil {
		logger.Errorf("msgid:%s,err:%v", msgId, err)
		err = status.Error(codes.Internal, "更新客户端系统配置失败！msgid:"+msgId)
	}
	return
}
