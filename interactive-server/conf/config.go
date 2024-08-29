package conf

import (
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type configParams struct {
	WebsocketServerPort string `yaml:"websocketServerPort"` //本地websocket服务端口
	GRPCServerAddr      string `yaml:"grpcServerAddr"`      //远程调用grpc服务地址
}

var Conf = func() (c configParams) {
	var confFile = "server.conf"
	yamlFile, err := os.ReadFile(confFile)
	if err != nil {
		log.Println("配置文件", confFile, " 不存在")
		return configParams{
			WebsocketServerPort: ":8558",
			GRPCServerAddr:      "localhost:8040",
		}
	}

	yaml.Unmarshal(yamlFile, &c)
	if !strings.HasPrefix(c.WebsocketServerPort, ":") {
		c.WebsocketServerPort = ":" + c.WebsocketServerPort
	}
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}()
