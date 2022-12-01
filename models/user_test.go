package model

import (
	"Rhine-Cloud-Driver/config"
	log "Rhine-Cloud-Driver/logic/log"
	"fmt"
	"io/ioutil"
	"testing"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func init() {
	var cf config.Config
	configFile, err := ioutil.ReadFile("../conf/Rhine-Cloud-Driver.yaml")
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
	Init(cf)
}

func Test_AddUser(t *testing.T) {
	user := User{
		Uid:      "1",
		Password: "1",
	}
	fmt.Printf("%v\n", setHaltHash(user.Password))
}
