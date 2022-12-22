package authenticator

import (
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/server/authenticator"
	"github.com/fatmalabidi/MailBoxServcie/src/storage"
	tdStorage "github.com/fatmalabidi/MailBoxServcie/src/testdata/storage"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
)

// TTLoadCredentials is a test table for the LoadCredentials function.
type TTLoadCredentials struct {
	Name     string
	Auth     *authenticator.Authenticator
	HasError bool
}

const (
	Tusername string = "93011917-bb2e-4e91-a7f4-65d52e88e189"
	Tpasword  string = "random-password"
)

// CreateTTLoadCredentials returns a test table for the LoadCredentials function.
func CreateTTLoadCredentials() []TTLoadCredentials {
	logger := helpers.GetTestLogger()
	conf, _ := config.LoadConfig()

	conf.Storage.CredentialsPath = tdStorage.CredentialsTestPath

	validS3handler, _ := storage.Create(&conf.Storage, logger)

	validAuth := authenticator.NewAuthenticator(validS3handler, &conf.Storage)

	// make invalid conf
	invalidCredPath, _ := config.LoadConfig()
	invalidCredPath.Storage.CredentialsPath = "invalid-path"

	//invalid handlers
	invalidS3handler, _ := storage.Create(&invalidCredPath.Storage, logger)

	invalidAuthKey := authenticator.NewAuthenticator(validS3handler, &invalidCredPath.Storage)
	invalidAuthCredPath := authenticator.NewAuthenticator(invalidS3handler, &invalidCredPath.Storage)

	return []TTLoadCredentials{
		{
			Name:     "valid auth",
			Auth:     validAuth,
			HasError: false,
		},
		{
			Name:     "invalid authenticator: invalid credPath",
			Auth:     invalidAuthCredPath,
			HasError: true,
		},
		{
			Name:     "invalid authenticator: invalid encryption key",
			Auth:     invalidAuthKey,
			HasError: true,
		},
	}
}
