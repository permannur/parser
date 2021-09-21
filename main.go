package main

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"parser/config"
	"parser/handler"
	"parser/logger"
	"syscall"
	"time"
)

func main() {
	var err error
	// init config params from config.json
	err = config.Init()
	if err != nil {
		log.Printf("main: connot init config err=%s\n", err)
		return
	}

	// logger - depend on config's debug param, writes or don't write logs
	logger.Init(os.Stdout)

	r := mux.NewRouter()
	r.HandleFunc("/sites", handler.Search).Methods("GET")
	srv := &http.Server{
		Handler: r,
		Addr:    config.Values().GetServer().GetHost() + ":" + config.Values().GetServer().GetPort(),
	}

	log.Print("listening on ", config.Values().GetServer().GetHost()+":"+config.Values().GetServer().GetPort())

	// other things for gracefully shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")
	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}
