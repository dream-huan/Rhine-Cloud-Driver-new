package jwt

import (
	log "Rhine-Cloud-Driver/logic/log"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"time"
)

// JWT个性化claims结构体
type CustomClaims struct {
	Uid uint64
	jwt.StandardClaims
}

var jwtkey string

func Init(key string) {
	jwtkey = key
}

func GenerateToken(uid uint64) (string, error) {
	expireTime := time.Now().Add(time.Second * 60 * 60 * 24 * 7) //登录有效期为7天
	claims := CustomClaims{
		Uid: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtkey))
	if err != nil {
		log.Logger.Error("JWT的token生成错误:%#v", zap.Error(err))
	}
	return token, err
}

func TokenValid(token string) bool {
	tokenClaims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtkey), nil
	})
	if err != nil {
		log.Logger.Error("JWT的token提取值错误:%#v", zap.Error(err))
		return false
	}
	if tokenClaims != nil {
		_, ok := tokenClaims.Claims.(*CustomClaims)
		//fmt.Printf("%#v %#v %#v %#v", ok, tokenClaims.Valid, claims, ip)
		//if ok && tokenClaims.Valid && claims.StandardClaims.Audience == ip {
		if ok && tokenClaims.Valid {
			return true
		}
	}
	return false
}

func TokenGetUid(token string) (isok bool, uid uint64) {
	if !TokenValid(token) {
		return false, 0
	}
	tokenClaims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtkey), nil
	})
	if err != nil {
		// 这里的错误100%是此token非法，有人伪造，直接判定为无效token
		return false, 0
	}
	claims, _ := tokenClaims.Claims.(*CustomClaims)
	return true, claims.Uid
}

func TokenGetIp(token string) (uid string) {
	tokenClaims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtkey), nil
	})
	if err != nil {
		return "0.0.0.0"
	}
	claims, _ := tokenClaims.Claims.(*CustomClaims)
	return claims.StandardClaims.Audience
}

func ParseToken(token string) (*CustomClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtkey), nil
	})
	if err != nil {
		return nil, err
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*CustomClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}
