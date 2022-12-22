package storage

import (
	"context"
	"errors"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/storage/s3"
	"github.com/sirupsen/logrus"
)

// Handler holds functions related to storage
type Handler interface {
	// Download downloads a file based on the filePath and returns a buffer
	Download(ctx context.Context, bucket string, filePath string) ([]byte, error)
}

// Create creates storage handler of type defined in config
func Create(storage *config.Storage, logger *logrus.Logger) (Handler, error) {
	var storageHandler Handler
	switch storage.Type {
	case "s3":
		storageHandler = s3.NewS3Handler(storage, logger)
	default:
		return nil, errors.New("invalid storage type")
	}
	return storageHandler, nil
}
