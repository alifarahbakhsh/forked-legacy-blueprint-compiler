package parser

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)


type Address struct {
	Name string `json:"name"`
	Address string `json:"address"`
	Port int `json:"port"`
	Hostname string `json:"hostname"`
}

type Node struct {
	Hostname string `json:"hostname"`
	// Opts map[string]bool `json:options`
	IsBuildNode bool `json:"is_build_node"`
}

type EnvVariable struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type Environment struct {
	Name string `json:"name"`
	Variables []EnvVariable `json:"variables"`
}
	
type Config struct {
	AppName string `json:"app_name"`
	SrcDir string `json:"src_dir"`
	OutDir string `json:"output_dir"`
	WiringFile string `json:"wiring_file"`
	Target string `json:"target"`
	Addresses []Address `json:"addresses"`
	Inventory []Node `json:"inventory"`
	Environment []Environment `json:"environment"`
}

func ParseConfig (filename string) (*Config, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var config Config
	json.Unmarshal(bytes, &config)
	config.AppName = strings.ToLower(config.AppName)
	return &config, nil
}