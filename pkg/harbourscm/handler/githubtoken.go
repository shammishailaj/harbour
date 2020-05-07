package handler

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/harbourrocks/harbour/pkg/harbourscm/redis"
	"github.com/harbourrocks/harbour/pkg/redisconfig"
	l "github.com/sirupsen/logrus"
	"time"
)

// GenerateGithubToken generates a new github access token.
// The token is generated by signing a self created jwt with the private key
// shared with github once the app was created.
// The private key was initially generated by github.
func GenerateGithubToken(appId int, validity time.Duration, redisOptions redisconfig.RedisOptions) (tokenString string, err error) {
	// resolve private key from redis
	client := redisconfig.OpenClient(redisOptions)
	keyBytes, err := client.HGet(redis.GithubAppKey(appId), "pem").Result()
	if err != nil {
		l.WithField("Lookup", "GithubApp").WithField("AppId", appId).WithError(err).Error()
		return
	}

	// log private key only when tracing
	l.WithField("PEM", keyBytes).Trace("Got pem key from redis")

	// required claims https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#accessing-api-endpoints-as-a-github-app
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(validity).Unix(),
		"iss": appId,
	}

	// keyBytes is in PEM format, convert to raw key bytes
	pKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyBytes))
	if err != nil {
		l.WithField("AppId", appId).WithError(err).Error("Failed to convert private key")
		return
	}

	// build the jwt
	// github requires RS256 https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#accessing-api-endpoints-as-a-github-app
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// sign the jwt
	tokenString, err = token.SignedString(pKey)
	if err != nil {
		l.WithField("AppId", appId).WithError(err).Error("Failed to sign jwt token")
		return
	}

	return
}
