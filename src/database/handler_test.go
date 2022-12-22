package database

import (
	db "github.com/fatmalabidi/MailBoxServcie/src/testdata/database"
	hlp "github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"testing"
)

func TestCreateMessagesHandler(t *testing.T) {
	tableTest := db.CreateTTMessagesHandler()
	log := hlp.GetTestLogger()
	for _, testCase := range tableTest {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := CreateMessagesHandler(&testCase.MessagesDB, log)
			if err != nil && !testCase.HasError {
				t.Errorf("expected success , got error: %v", err)
			}
			if err == nil && testCase.HasError {
				t.Error("expected error, got nil")
			}
		})
	}
}
