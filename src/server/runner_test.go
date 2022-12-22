package server_test

import (
	"github.com/fatmalabidi/MailBoxServcie/src/server"
	"github.com/fatmalabidi/MailBoxServcie/src/testdata"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"testing"
)

func TestCreate(t *testing.T) {
	logger := helpers.GetTestLogger()
	testCases := testdata.PrepareCreateTestData()
	for _, testcase := range testCases {
		t.Run(testcase.Name, func(t *testing.T) {
			_, err := server.Create(testcase.Config, logger)
			if testcase.HasError && err == nil {
				t.Error("expected to get error, got nil")
			}
			if !testcase.HasError && err != nil {
				t.Errorf("expected success, got error %v", err)
			}
			// TODO add expected error check after adding customized errors
		})
	}

}
