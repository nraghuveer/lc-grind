package lc_api

import (
	"io/ioutil"
	"log"
	"sync"

	"gopkg.in/yaml.v3"
)

var once sync.Once
var config *lcConfig

type lcConfig struct {
	CSRF       string
	LC_SESSION string
}

func (config *lcConfig) load() *lcConfig {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}
	data := make(map[string]string)
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		log.Fatalf("Failed to Unmarshal the config file: %v", err)
	}
	config.LC_SESSION = data["LC_SESSION"]
	config.CSRF = data["CSRF"]
	return config
}

func GetLcConfig() *lcConfig {
	config := &lcConfig{}
	config.load()
	return config
}
