package authenticator

import (
	"context"
	"encoding/json"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/storage"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
)

// Authenticator struct is responsible of loading the credentials from storage, decrypt them using encryption and inject them to environment
type Authenticator struct {
	Storage       storage.Handler
	StorageConfig *config.Storage
}

// NewAuthenticator is the factory function of Authenticator entity
func NewAuthenticator(storage storage.Handler, storageConfig *config.Storage) *Authenticator {
	return &Authenticator{
		Storage:       storage,
		StorageConfig: storageConfig,
	}
}

// TODO remove the decryption part after configuring S3 to do the encryption/decryption
// LoadCredentials function receives a storage and an encryption handler and uses them to:
// * download the credentials file from storage
// * inject them to the environment
func (auth *Authenticator) LoadCredentials(ctx context.Context) error {
	// download credentials
	buffer, err := auth.Storage.Download(ctx, auth.StorageConfig.Bucket, auth.StorageConfig.CredentialsPath)
	if err != nil {
		return err
	}
	// get encrypted and base64 encoded credentials
	username, password, err := getEncryptedCredentials(buffer)
	if err != nil {
		return err
	}

	return helpers.SetCredToEnv(username, password)
}

// getEncryptedCredentials function unmarshal a buffer to a credentials' struct and returns the username and password
func getEncryptedCredentials(buffer []byte) (string, string, error) {
	cred := struct {
		Username string
		Password string
	}{}
	err := json.Unmarshal(buffer, &cred)
	if err != nil {
		return "", "", err
	}

	return cred.Username, cred.Password, nil
}
