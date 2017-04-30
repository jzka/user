package main

//TODO: Add prometheus and distributed logging
//TODO: HATEOAS

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	corelog "log"

	"context"

	"github.com/go-kit/kit/log"
	"github.com/user"
	db "github.com/user/dbOperations"
)

var (
	port string
)

const (
	ServiceName = "user"
)

func init() {
	flag.StringVar(&port, "port", "8084", "Port on which to run")
}

func main() {

	flag.Parse()
	// Mechanical stuff.
	errc := make(chan error)
	ctx := context.Background()

	// Log domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
	}
	dbconn := false
	dbm := db.Mongo{}
	for !dbconn {
		err := dbm.Init()
		if err != nil {
			corelog.Print(err)
		} else {
			dbconn = true
		}
	}
	var svc user.Service
	svc = user.NewUserService(&dbm, logger)
	endpoints := user.MakeEndpoints(svc)
	router := user.MakeHTTPHandler(ctx, endpoints, logger)
	// Create and launch the HTTP server.
	go func() {
		logger.Log("transport", "HTTP", "port")
		errc <- http.ListenAndServe(fmt.Sprintf(":%v", port), router)
	}()

	// Capture interrupts.
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("exit", <-errc)
}
