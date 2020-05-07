package harbouriam

import (
	"fmt"
	"github.com/harbourrocks/harbour/pkg/harbouriam/configuration"
	"github.com/harbourrocks/harbour/pkg/harbouriam/handler"
	"github.com/harbourrocks/harbour/pkg/httphandler"
	"github.com/harbourrocks/harbour/pkg/httphandler/traits"
	"github.com/harbourrocks/harbour/pkg/redisconfig"
	"net/http"
	"net/url"
	"path"

	"github.com/sirupsen/logrus"
)

// RunIAMServer runs the IAM server application
func RunIAMServer(o *configuration.Options) error {
	logrus.Info("Started Harbour IAM server")

	// obtain login redirect url
	redirectURL, err := url.Parse(o.IAMBaseURL)
	if err != nil {
		logrus.Fatal(err)
	} else {
		redirectURL.Path = path.Join(redirectURL.Path, "/auth/oidc/callback")
	}

	http.HandleFunc("/auth/test", func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace(r)
		model := handler.AuthModel{}
		traits.AddHttp(&model, r, w, o.OIDCConfig)
		traits.AddIdToken(&model)

		model.Handle()
	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace(r)
		model := handler.ProfileModel{}
		traits.AddHttp(&model, r, w, o.OIDCConfig)
		traits.AddIdToken(&model)
		redisconfig.AddRedis(&model, o.Redis)

		if err := httphandler.ForceAuthenticated(&model); err != nil {
			model.Handle()
		}
	})

	// DockerHandler
	http.HandleFunc("/docker/password", func(w http.ResponseWriter, r *http.Request) {
		logrus.Trace(r)
		model := handler.DockerModel{}
		traits.AddHttp(&model, r, w, o.OIDCConfig)
		traits.AddIdToken(&model)
		redisconfig.AddRedis(&model, o.Redis)

		if err := httphandler.ForceAuthenticated(&model); err != nil {
			model.Handle()
		}
	})

	bindAddress := "127.0.0.1:5100"
	logrus.Info(fmt.Sprintf("Listening on httphandler://%s/", bindAddress))

	err = http.ListenAndServe(bindAddress, nil)
	logrus.Fatal(err)

	return err
}
