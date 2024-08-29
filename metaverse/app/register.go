package app

import (
	"log"
	"metaverse/conf"
	"metaverse/pbs"
	"metaverse/servers"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func register() *grpc.Server {
	s := grpc.NewServer(grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
		grpc.ChainUnaryInterceptor(unaryEndToEndInterceptor, unaryTokenCheckInterceptor))
	pbs.RegisterUserServiceServer(s, &servers.UserServer{})
	pbs.RegisterConfigServiceServer(s, &servers.ConfigServer{})
	pbs.RegisterRoleServiceServer(s, &servers.RoleServer{})
	return s
}

func Run() {
	lis, err := net.Listen("tcp", conf.Conf.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := register()
	reflection.Register(s)
	log.Println("启动grpc服务成功，端口: ", conf.Conf.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
