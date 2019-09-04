package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func Run(configFilePath string, isDaemon bool) error {
	if isDaemon {
		return runDaemon()
	}
	config, err := getConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	for i := range config {
		wg.Add(1)
		conf := config[i]
		go func() {
			err = runServer(&wg, conf)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()
	time.Sleep(10 * time.Second)

	os.Exit(0)
	return nil
}

func runServer(wg *sync.WaitGroup, configServer balancerConfig) error {
	defer wg.Done()

	srv := New(&http.Server{
		Addr:         configServer.NetworkInterface,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	})

	for i := range configServer.Upstreams {
		item := i
		switch proxyMethod := configServer.Upstreams[item].ProxyMethod; proxyMethod {
		case "anycast":

			srv.HandleFunc(
				configServer.Upstreams[item].HTTPMethods,
				configServer.Upstreams[item].HTTPPath,
				func(writer http.ResponseWriter, request *http.Request) {
					anycast(writer,
						request,
						configServer.Upstreams[item].Backends)
				})

		case "round-robin":
			rr := NewRoundRobin()

			srv.HandleFunc(
				configServer.Upstreams[item].HTTPMethods,
				configServer.Upstreams[item].HTTPPath,
				func(writer http.ResponseWriter, request *http.Request) {
					roundRobin(
						writer,
						request,
						configServer.Upstreams[item].Backends,
						rr)
				})
		}
	}
	log.Println("starting server", configServer)

	return srv.ListenAndServe()
}

func StopServer() error {
	if _, err := os.Stat(getPidFilePath()); err != nil {
		return ErrNotRunning
	}

	data, err := ioutil.ReadFile(getPidFilePath())
	if err != nil {
		return ErrNotRunning
	}
	ProcessID, err := strconv.Atoi(string(data))

	if err != nil {
		return ErrUnableToParse
	}

	process, err := os.FindProcess(ProcessID)
	if err != nil {
		return ErrUnableToParse
	}
	// remove PID file
	os.Remove(getPidFilePath())

	fmt.Printf("Sending SIGTERM to process ID [%v] now.\n", ProcessID)
	// kill process and exit immediately
	log.Println(ProcessID)
	err = process.Signal(syscall.SIGTERM)

	if err != nil {
		log.Println(err)
		return ErrUnableToKill
	}

	fmt.Printf("Sending SIGTERM to process ID [%v]\n", ProcessID)
	return nil
}

func ReloadServer() error {
	err := StopServer()
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	err = runDaemon()
	if err != nil {
		return err
	}
	return nil
}
