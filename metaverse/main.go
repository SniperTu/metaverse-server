package main

import (
	"metaverse/app"
	server "metaverse/app/servers"
)

func main() {
	server.RunHTTPServer()
	app.Run()
}
