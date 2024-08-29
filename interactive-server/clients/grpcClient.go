package clients

import (
	"interactive-server/conf"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var GrpcConn = func() *grpc.ClientConn {
	conn, err := grpc.NewClient(conf.Conf.GRPCServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect:%v", err)
	}
	return conn
}()
