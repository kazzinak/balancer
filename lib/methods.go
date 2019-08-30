package lib

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
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

func balancerAnycast(writer http.ResponseWriter, request *http.Request, backends []backend) {

	ctx, cancel := context.WithCancel(context.Background())
	anycastCh := make(chan []byte)
	for _, url := range backends {
		go func(url backend, ctx context.Context) {
			proxyReq, err := http.NewRequest(request.Method, string(url), request.Body)

			client := http.Client{}

			resp, err := client.Do(proxyReq)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()

			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case anycastCh <- []byte(respBody):

			}
		}(url, ctx)
	}
	select {
	case respBody := <-anycastCh:
		writer.Write([]byte(respBody))
		cancel()
	}

}

func balancerRoundRobin(writer http.ResponseWriter, request *http.Request, backends []backend, r *RoundRobin) {
	r.backends = backends

	log.Println(r.current)
	url := r.Get()
	log.Println(r.current)

	fmt.Fprintf(writer, string(url))

	proxyReq, err := http.NewRequest(request.Method, string(url), request.Body)

	client := http.Client{}

	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	writer.Write([]byte(respBody))
}
