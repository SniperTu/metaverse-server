package main

import (
	"context"
	"flag"
	"fmt"
	"interactive-server/conf"
	"interactive-server/services"
	"interactive-server/utils"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	addr = flag.String("addr", ":8443", "http service address")
)

func main() {
	flag.Parse()
	serv := new(services.GameService)
	if flag.NArg() < 0 {
		log.Fatal("filename not specified")
	}
	mu := http.NewServeMux()
	mu.HandleFunc("/interactive", serv.Start)
	mu.HandleFunc("/banNotify/", services.BanNotifyHandler)
	uurl, _ := url.Parse(fmt.Sprintf("http://localhost:%d/", utils.PProf()))
	mu.HandleFunc("/debug/pprof/", func(w http.ResponseWriter, r *http.Request) {
		httputil.NewSingleHostReverseProxy(uurl).ServeHTTP(w, r)
	})
	srv := &http.Server{
		Addr:    conf.Conf.WebsocketServerPort,
		Handler: mu,
	}
	lc := net.ListenConfig{KeepAlive: -1}
	ln, err := lc.Listen(context.Background(), "tcp", conf.Conf.WebsocketServerPort)
	if err != nil {
		log.Fatal(err)
	}
	if err = srv.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
