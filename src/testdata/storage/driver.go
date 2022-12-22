package storage

import (
	"github.com/fatmalabidi/MailBoxServcie/src/config"
)

const CredentialsTestPath string = "test/config-test.json"

// TTDownload is a test table for the Download function.
type TTDownload struct {
	Name     string
	Bucket   string
	ItemPath string
	Timeout  int64
	HasError bool
}

// CreateTTDownload returns a test table for the Download function
func CreateTTDownload() []TTDownload {
	conf, _ := config.LoadConfig()
	return []TTDownload{
		{
			Name:     "valid context, valid bucket and valid CredentialsPath",
			Bucket:   conf.Storage.Bucket,
			Timeout:  10,
			ItemPath: CredentialsTestPath,
			HasError: false,
		},
		{
			Name:     "invalid Bucket",
			Bucket:   "non-sense-bucket-name",
			ItemPath: CredentialsTestPath,
			Timeout:  10,
			HasError: true,
		},
		{
			Name:     "invalid CredentialsPath",
			Bucket:   conf.Storage.Bucket,
			ItemPath: "non-sense-ItemPath-name",
			Timeout:  10,
			HasError: true,
		},
	}
}
