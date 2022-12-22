package s3_test

import (
	"context"
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/storage/s3"
	td "github.com/fatmalabidi/MailBoxServcie/src/testdata/storage"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"testing"
)

func Test_Download(t *testing.T) {
	logger := helpers.GetTestLogger()
	conf, _ := config.LoadConfig()
	handler := s3.NewS3Handler(&conf.Storage, logger)
	for _, testCase := range td.CreateTTDownload() {
		t.Run(testCase.Name, func(t *testing.T) {

			ctx, cancel := helpers.AddTimeoutToCtx(context.Background(), testCase.Timeout)
			defer cancel()
			buffer, err := handler.Download(ctx, testCase.Bucket, testCase.ItemPath)
			if err == nil && testCase.HasError {
				t.Errorf("expected error, got data %v", buffer)
			}
			if (err != nil || len(buffer) == 0) && (!testCase.HasError) {
				t.Errorf("expected success, got %v", err)
			}
		})
	}
}
