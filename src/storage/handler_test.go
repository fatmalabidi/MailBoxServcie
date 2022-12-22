package storage_test

import (
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/storage"
	td "github.com/fatmalabidi/MailBoxServcie/src/testdata/storage"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"testing"
)

func Test_Create(t *testing.T) {
	for _, testCase := range td.TTCreateS3Handler() {
		conf := config.Config{
			Storage: config.Storage{
				Type: testCase.Type,
			},
		}
		log := helpers.GetTestLogger()
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := storage.Create(&conf.Storage, log)
			if err != nil && !testCase.HasError {
				t.Errorf("expected success , got error: %v", err)
			}
			if err == nil && testCase.HasError {
				t.Error("expected error")
			}
		})
	}
}
