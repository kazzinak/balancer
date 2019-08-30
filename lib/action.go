package lib

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func Run(configFilePath string) error {
	config, err := getConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	for i := range config {
		conf := config[i]
		go func() {
			err = runBalancer(ctx, &wg, conf)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	<-c
	fmt.Println("cancel")
	cancel()

	wg.Wait()
	time.Sleep(10 * time.Second)

	return nil
}

func runBalancer(ctx context.Context, wg *sync.WaitGroup, configServer balancerConfig) error {
	defer wg.Done()

	r := mux.NewRouter()
	log.Println(configServer.NetworkInterface)

	srv := &http.Server{
		Addr:         configServer.NetworkInterface,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	for i := range configServer.Upstreams {
		item := i
		if configServer.Upstreams[item].ProxyMethod == "anycast" {
			r.HandleFunc(configServer.Upstreams[item].HTTPPath, func(writer http.ResponseWriter, request *http.Request) {
				balancerAnycast(writer, request, configServer.Upstreams[item].Backends)
			})
		} else if configServer.Upstreams[item].ProxyMethod == "round-robin" {
			rr := NewRoundRobin()
			r.HandleFunc(configServer.Upstreams[item].HTTPPath, func(writer http.ResponseWriter, request *http.Request) {
				balancerRoundRobin(writer, request, configServer.Upstreams[item].Backends, rr)
			})
		}
	}

	fmt.Println("starting balancer server")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err.Error())
		}
	}()

	<-ctx.Done()

	log.Println("shutting down")

	ctxInternal, cancelInternal := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancelInternal()

	return srv.Shutdown(ctxInternal)
}
