package lib

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type RoundRobin struct {
	sync.Mutex
	current  int
	backends []backend
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		current:  0,
		backends: []backend{},
	}
}

func (r *RoundRobin) Get() backend {
	r.Lock()
	defer r.Unlock()

	if r.current >= len(r.backends) {
		r.current = r.current % len(r.backends)
	}

	backend := r.backends[r.current]
	r.current++
	return backend
}

func anycast(writer http.ResponseWriter, request *http.Request, backends []backend) {
	retries := 1

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	anycastCh := make(chan []byte)

	for i := 0; i < retries; i++ {
		for _, url := range backends {
			go func(url backend, ctx context.Context) {

				respBody, err := requester(request, url)
				if err != nil {
					log.Println(err)
				}

				select {
				case <-ctx.Done():
					return
				case anycastCh <- []byte(respBody):
				}
			}(url, ctx)
		}
	}
	select {
	case respBody := <-anycastCh:
		writer.Write([]byte(respBody))
		cancel()
	}

}

func roundRobin(writer http.ResponseWriter, request *http.Request, backends []backend, r *RoundRobin) {
	r.backends = backends
	var respBody []byte
	var err error

	for range backends {
		url := r.Get()
		fmt.Fprintf(writer, string(url))
		respBody, err = requester(request, url)
		if err != nil {
			log.Println(err)
		} else {
			break
		}
	}
	_, err = writer.Write(respBody)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
	}
}

func requester(request *http.Request, url backend) ([]byte, error) {

	proxyReq, err := http.NewRequest(request.Method, string(url), request.Body)

	client := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Println(err)
		return []byte(""), err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	return []byte(respBody), nil
}
