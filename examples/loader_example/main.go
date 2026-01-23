package main

import (
	"log"
	"net/http"
	"os"

	"github.com/brnsampson/ezconf/httpconf"
	"github.com/brnsampson/optional"
)

type MyServiceConfig struct {
	Name         string
	Description  string
	NodeID       uint32 `flag:"true" required:"true" field:"node"`
	Priority     uint16
	SecretKey    optional.Secret `flag:"true" default:"secretkey.txt"`
	ServerConfig httpconf.HttpServerConfig
}

type MyDBConfig struct {
	Address string `flag:"true" default:"127.0.0.1"`
	Port    uint16 `flag:"true" default:"8080"`
}

//go:generate ezconf -path=/etc/myapp/ -flagDefault=false
type MyAppConfig struct {
	MyService MyServiceConfig
	MyDB      MyDBConfig
}

func main() {
	l, err := NewLoader()
	if err != nil {
		// Uh-oh!
		os.Exit(1)
	}

	conf := l.Prev()
	log.Println("Starting server for Name: ", conf.MyService.Name, " NodeID: ", conf.MyService.NodeID)

	myServer := conf.MyService.ServerConfig.NewHttpServer()
	err = myServer.ListenAndServe() // Listen until interrupted
	if err == http.ErrServerClosed {
		// Exited due to myServer.Shutdown or myServer.Close() being called.
		os.Exit(0)
	}

	os.Exit(2)
}
