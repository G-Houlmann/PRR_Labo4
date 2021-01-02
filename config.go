package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Topology struct {
	Debug       bool     `json:"debug"`
	Trace       bool     `json:"trace"`
	ClientCount int      `json:"clientCount"`
	Clients     []Client `json:"clients"`
}

type Client struct {
	Id           int    `json:"id"`
	Hostname     string `json:"hostname"`
	Port         int    `json:"port"`
	PrimeDivisor int    `json:"primeDivisor"`
	Neighbors    []int  `json:"neighbors"`
}

//Parse the config.json file in a Topology struct
func Parse(path string) (*Topology, error) {
	jsonFile, err := os.Open(path)

	if err != nil {
		//fmt.Println(err)
		return nil, err
	}

	var topology Topology

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(byteValue, &topology)
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}

	defer jsonFile.Close()

	return &topology, nil
}
