package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type balancerConfig struct {
	NetworkInterface string     `json:"interface"`
	Upstreams        []upstream `json:"upstreams`
}

type upstream struct {
	HTTPPath    string    `json:"path"`
	HTTPMethods string    `json:"methods"`
	Backends    []backend `json:"backends"`
	ProxyMethod string    `json:"proxyMethod"`
}

type backend string

// var config []balancerConfig

func configParser(configFile *os.File) ([]balancerConfig, error) {

	config := []balancerConfig{}

	b, err := ioutil.ReadAll(configFile)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		return config, err
	}

	fmt.Println(config)

	return config, nil
}

func getConfig(configFilePath string) ([]balancerConfig, error) {

	config := []balancerConfig{}

	if configFilePath == "" {
		configFilePath = "config.json"
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	defer configFile.Close()

	config, err = configParser(configFile)
	if err != nil {
		log.Fatal(err)
	}
	return config, nil
}
