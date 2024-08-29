package servers

import (
	"context"
	"errors"
	server "metaverse/app/servers"
	"metaverse/models"
	"metaverse/pbs"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleServer struct {
	server.Server
}

func (R *RoleServer) Create(ctx context.Context, req *pbs.RoleReq) (resp *pbs.Empty, err error) {
	resp = new(pbs.Empty)
	loginUser, err := R.Auth(ctx)
	if err != nil {
		return
	}
	userInfo, err := new(models.User).GetUserById(loginUser.Id)
	if err != nil {
		return
	}
	if req.Name == "" {
		err = errors.New("请输入角色名称")
		return
	}
	if req.Mode == -1 {
		err = errors.New("必须设置模型")
		return
	}
	role := &pbs.Role{}
	role.CreatedAt = time.Now().Unix()
	role.Id = primitive.NewObjectID().Hex() //创建ID
	if req.Avatar != "" {
		role.Avatar = req.Avatar
	}
	if req.Gender != -1 {
		role.Gender = req.Gender
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Mode != -1 {
		role.Mode = req.Mode
	}
	userInfo.Roles = append(userInfo.Roles, role)
	err = new(models.User).Update(userInfo)
	return
}

func (R *RoleServer) Modify(ctx context.Context, req *pbs.RoleReq) (resp *pbs.Role, err error) {
	loginUser, err := R.Auth(ctx)
	if err != nil {
		return
	}
	userInfo, err := new(models.User).GetUserById(loginUser.Id)
	if err != nil {
		return
	}
	role := &pbs.Role{}
	updated := false //是否修改
	if len(userInfo.Roles) > 0 {
		for i, r := range userInfo.Roles {
			if r.Id == req.Id {
				if req.Avatar != "" { //头像
					r.Avatar = req.Avatar
				}
				if req.Gender != -1 { //性别
					r.Gender = req.Gender
				}
				if req.Name != "" { //名称
					r.Name = req.Name
				}
				if req.Mode != -1 { //模型
					r.Mode = req.Mode
				}
				userInfo.Roles[i] = r
				role = r
				updated = true
			}
		}
	}
	if !updated {
		if req.Name == "" {
			err = errors.New("请输入角色名称")
			return
		}
		if req.Mode == -1 {
			err = errors.New("必须设置模型")
			return
		}
		role.CreatedAt = time.Now().Unix()
		role.Id = primitive.NewObjectID().Hex() //创建ID
		if req.Avatar != "" {
			role.Avatar = req.Avatar
		}
		if req.Gender != -1 {
			role.Gender = req.Gender
		}
		if req.Name != "" {
			role.Name = req.Name
		}
		if req.Mode != -1 {
			role.Mode = req.Mode
		}
		userInfo.Roles = append(userInfo.Roles, role)
	}
	err = new(models.User).Update(userInfo)
	if err == nil {
		resp = role
	}
	return
}
