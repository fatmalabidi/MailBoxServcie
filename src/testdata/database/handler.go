package database

import "github.com/fatmalabidi/MailBoxServcie/src/config"

// TTCreateHandler represents table test structure of CreateHandler test
type TTCreateHandler struct {
	Name         string
	DatabaseConf config.Database
	HasError     bool
}

// TTCreateMessagesHandler represents table test structure of CreateMessagesHandler test
type TTCreateMessagesHandler struct {
	Name       string
	MessagesDB config.Messages
	HasError   bool
}

// CreateTTMessagesHandler creates table test for TTCreateMessagesHandler test
func CreateTTMessagesHandler() []TTCreateMessagesHandler {
	testTableName := "dynamodb"
	return []TTCreateMessagesHandler{
		{
			Name:       "valid config",
			MessagesDB: config.Messages{Type: testTableName},
			HasError:   false,
		},
		{
			Name:       "unsupported db ",
			MessagesDB: config.Messages{Type: "unknown"},
			HasError:   true,
		},
	}
}
