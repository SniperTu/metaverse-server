package conf

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type configFile struct {
	Port string `yaml:"port"`
	Db   struct {
		Mongo struct {
			Hosts    []string `yaml:"hosts"`
			User     string   `yaml:"user"`
			Pwd      string   `yaml:"pwd"`
			Database string   `yaml:"database"`
		} `yaml:"mongo"`
		Redis struct {
			Host     string `yaml:"host"`
			Pwd      string `yaml:"pwd"`
			Database int    `yaml:"database"`
		} `yaml:"redis"`
	} `yaml:"db"`
	InteractiveServerHTTPAddr string `yaml:"interactiveServerHTTPAddr"`
	HTTPPort                  string `yaml:"httpPort"`
}

var Conf = func() (c configFile) {
	var confFile = "server.conf"
	yamlFile, err := os.ReadFile(confFile)
	if err != nil {
		log.Println("配置文件", confFile, " 不存在")
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}()
