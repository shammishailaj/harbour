package app

import (
	"github.com/harbourrocks/harbour/pkg/harbouriam"
	"github.com/harbourrocks/harbour/pkg/logconfig"
	"github.com/harbourrocks/harbour/pkg/redisconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewIAMServerCommmand creates a *cobra.Command object with default parameters
func NewIAMServerCommmand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "harbour-iam",
		Long: `The harbour.rocks IAM server manages
authentication and authorization for the harbour environment.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			// load OIDC  config
			s := harbouriam.ParseViperConfig()

			// load redis config
			s.Redis = redisconfig.ParseViperConfig()

			// configure logging
			l := logconfig.ParseViperConfig()
			logconfig.ConfigureLog(l)

			logrus.Info("Harbour IAM configured")

			// test redis connection
			redisconfig.TestConnection(s.Redis)

			return harbouriam.RunIAMServer(s)
		},
	}

	return cmd
}

func init() {
	cobra.OnInitialize(initCobra)
}

func initCobra() {
	viper.AutomaticEnv()
}
