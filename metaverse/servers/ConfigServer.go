package servers

import (
	"context"
	server "metaverse/app/servers"
	"metaverse/models"
	"metaverse/pbs"
)

type ConfigServer struct {
	server.Server
}

func (C *ConfigServer) Update(ctx context.Context, req *pbs.Config) (resp *pbs.Empty, err error) {
	resp = new(pbs.Empty)
	_, err = C.Auth(ctx)
	if err != nil {
		return
	}
	err = new(models.Config).Update(req)
	return
}
