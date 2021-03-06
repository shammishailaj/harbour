package configuration

import (
	"github.com/harbourrocks/harbour/pkg/auth"
	"github.com/harbourrocks/harbour/pkg/redisconfig"
	l "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

// DockerOptions hold the options required to authenticate docker cli clients
type DockerOptions struct {
	SigningKeyPath     string
	CertificatePath    string
	ValidDockerService string
	Issuer             string
	TokenLifetime      time.Duration
}

// Options defines all options available to configure the IAM server.
type Options struct {
	OIDCClientID     string
	OIDCClientSecret string
	OIDCURL          string
	IAMBaseURL       string
	Redis            redisconfig.RedisOptions
	OIDCConfig       auth.OIDCConfig
	Docker           DockerOptions
}

// NewDefaultOptions returns the default options
func NewDefaultOptions() *Options {
	s := Options{
		OIDCClientID:     "",
		OIDCClientSecret: "",
		OIDCURL:          "",
		IAMBaseURL:       "",
		Redis:            redisconfig.NewDefaultRedisOptions(),
		OIDCConfig:       auth.DefaultConfig(),
	}

	return &s
}

// ParseViperConfig tries to map a viper configuration
func ParseViperConfig() *Options {
	s := NewDefaultOptions()

	s.OIDCClientID = viper.GetString("OIDC_CLIENT_ID")
	s.OIDCClientSecret = viper.GetString("OIDC_CLIENT_SECRET")
	s.OIDCURL = viper.GetString("OIDC_URL")
	s.IAMBaseURL = viper.GetString("IAM_BASE_URL")

	s.OIDCConfig = auth.ParseViperConfig()
	s.Redis = redisconfig.ParseViperConfig()

	s.Docker = DockerOptions{
		SigningKeyPath:     viper.GetString("DOCKER_TOKEN_SIGNING_KEY"),
		CertificatePath:    viper.GetString("DOCKER_TOKEN_CERTIFICATE"),
		ValidDockerService: viper.GetString("DOCKER_VALID_SERVICE"),
		Issuer:             viper.GetString("DOCKER_TOKEN_ISSUER"),
		TokenLifetime:      viper.GetDuration("DOCKER_TOKEN_LIFETIME"),
	}

	if _, err := os.Stat(s.Docker.SigningKeyPath); os.IsNotExist(err) {
		l.WithError(err).Fatal("Docker private key not found")
	}

	return s
}
