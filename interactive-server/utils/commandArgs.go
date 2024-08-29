package utils

import (
	"fmt"
	"os"
)

var (
	GitInfo    string //编译代码git提交信息:<branch-name><commit-id><commit-message>,编译时通过ldflags传入
	BuildTime  string //编译时间: time.RFC3339格式,编译时通过ldflags传入
	GoVersion  string //编译环境Go版本,编译时通过ldflags传入
	ServerName string //服务名称,编译时通过ldflags传入
)

var usageTemplateStr = fmt.Sprintf("Usage: \n\n\t"+
	"%s <commmand> [arguments]\n\n"+
	"The commmands are:\n\n"+
	"\t-v,--version\tprint server version info\n", ServerName)

func init() {
	args := os.Args
	if len(args) < 2 {
		return
	}
	switch arg1 := args[1]; arg1 {
	case "-v,--version":
		fmt.Printf("Server name: %s \n", ServerName)
		fmt.Printf("Git Commit Info: %s \n", GitInfo)
		fmt.Println()
		fmt.Printf("Build Time: %s \n", BuildTime)
		fmt.Printf("Go Version: %s \n", GoVersion)
		os.Exit(0)
	case "-h", "--help":
		fmt.Printf("Server name: %s \n", ServerName)
		fmt.Println(usageTemplateStr)
		os.Exit(0)
	default:
	}
	return
}
