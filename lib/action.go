package lib

import (
	"log"
	"net/http"
	"sync"
	"time"
)

func Run(configFilePath string) error {
	config, err := getConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// c := make(chan os.Signal, 1)

	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	for i := range config {
		wg.Add(1)
		conf := config[i]
		go func() {
			err = runBalancer(&wg, conf)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()
	time.Sleep(10 * time.Second)

	return nil
}

func runBalancer(wg *sync.WaitGroup, configServer balancerConfig) error {
	defer wg.Done()

	srv := New(&http.Server{
		Addr:         configServer.NetworkInterface,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	})

	for i := range configServer.Upstreams {
		item := i
		if configServer.Upstreams[item].ProxyMethod == "anycast" {
			srv.HandleFunc(configServer.Upstreams[item].HTTPMethods, configServer.Upstreams[item].HTTPPath, func(writer http.ResponseWriter, request *http.Request) {
				balancerAnycast(writer, request, configServer.Upstreams[item].Backends)
			})
		} else if configServer.Upstreams[item].ProxyMethod == "round-robin" {
			rr := NewRoundRobin()
			srv.HandleFunc(configServer.Upstreams[item].HTTPMethods, configServer.Upstreams[item].HTTPPath, func(writer http.ResponseWriter, request *http.Request) {
				balancerRoundRobin(writer, request, configServer.Upstreams[item].Backends, rr)
			})
		}
	}
	log.Println("starting server", configServer)

	return srv.ListenAndServe()
}
