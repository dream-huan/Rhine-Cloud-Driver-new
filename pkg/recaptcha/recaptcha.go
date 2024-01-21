package recaptcha

import (
	"Rhine-Cloud-Driver/pkg/log"
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
)

// google验证码回调结构体
type recaptchaToken struct {
	Success      bool    `json:"success"`
	Challenge_ts string  `json:"challenge_ts"`
	Hostname     string  `json:"hostname"`
	Score        float64 `json:"score"`
	Action       string  `json:"action"`
}

const recaptchaServerName = "https://recaptcha.google.cn/recaptcha/api/siteverify"

var privatekey string

func InitRecaptcha(key string) {
	privatekey = key
}

func VerifyToken(token string) bool {
	log.Logger.Info("recaptcha发送报文", zap.Any("url", recaptchaServerName), zap.Any("secret", privatekey), zap.Any("response", token))
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
	log.Logger.Info("recaptcha返回报文", zap.Any("body", resp.Body))
	var result recaptchaToken
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Logger.Error("对recaptcha结果处理错误:%#v", zap.Error(err))
	}
	log.Logger.Info("recaptcha返回结果", zap.Any("result", result))
	return result.Score >= 0.6
}
