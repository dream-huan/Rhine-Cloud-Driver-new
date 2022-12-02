package main

import (
	"Rhine-Cloud-Driver/config"
	log "Rhine-Cloud-Driver/logic/log"
	model "Rhine-Cloud-Driver/models"
	"Rhine-Cloud-Driver/routers"
	"fmt"
	"io/ioutil"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var cf config.Config

func InitConfig() {
	configFile, err := ioutil.ReadFile("./conf/Rhine-Cloud-Driver.yaml")
	if err != nil {
		fmt.Printf("%v", err)
		panic(err)
	}
	err = yaml.Unmarshal(configFile, &cf)
	if err != nil {
		fmt.Printf("%v", err)
		panic(err)
	}
	log.Logger, err = log.NewLogger(cf.Log.LogPath, cf.Log.LogLevel, cf.Log.MaxSize, cf.Log.MaxBackup,
		cf.Log.MaxAge, cf.Log.Compress, cf.Log.LogConsole, cf.Log.ServiceName)
	if err != nil {
		log.Logger.Error("Unmarshal yaml file error", zap.Error(err))
	}
	model.Init(cf)
}

func main() {
	InitConfig()
	routers.InitRouter(cf)
}
