package s3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	*s3manager.Downloader
	config *config.Storage
	log    *logrus.Logger
}

// NewS3Handler creates an AWS S3 Client and creates a s3Handler with the config needed
func NewS3Handler(config *config.Storage, logger *logrus.Logger) *Handler {
	// Create a single AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &Handler{
		Downloader: s3manager.NewDownloader(sess),
		config:     config,
		log:        logger,
	}
}

// Download downloads a file based on the filePath and returns a buffer
func (handler *Handler) Download(ctx context.Context, bucket string, filePath string) ([]byte, error) {

	var buffer aws.WriteAtBuffer
	_, err := handler.Downloader.DownloadWithContext(ctx, &buffer,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filePath),
		})
	return buffer.Bytes(), err
}
