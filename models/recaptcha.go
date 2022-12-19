package model

import (
	"Rhine-Cloud-Driver/logic/log"
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
)

// 谷歌验证码用于验证
const recaptchaServerName = "https://recaptcha.net/recaptcha/api/siteverify"

// google验证码回调结构体
type RecaptchaToken struct {
	Success      bool   `json:"success"`
	Challenge_ts string `json:"challenge_ts"`
	Hostname     string `json:"hostname"`
}

var privatekey string

func InitRecaptcha(key string) {
	privatekey = key
}

func VerifyToken(token string) bool {
	resp, err := http.PostForm(recaptchaServerName,
		url.Values{"secret": {privatekey}, "response": {token}})
	if err != nil {
		log.Logger.Error("httppost错误:%#v", zap.Error(err))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Error("ioutil的read方法错误:%#v", zap.Error(err))
	}
	var result RecaptchaToken
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Logger.Error("对recaptcha结果处理错误:%#v", zap.Error(err))
	}
	return result.Success
}
