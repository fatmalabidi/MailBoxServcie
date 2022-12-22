package database

import (
	"context"
	"fmt"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/database/dynamo"
	mdl "github.com/fatmalabidi/MailBoxServcie/src/models"
	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/sirupsen/logrus"
)

// MessagesHandler is the handler for the messages table
type MessagesHandler interface {
	// SendMessage adds the sent message to messages table
	SendMessage(ctx context.Context, req *svc.SendMessageReq, ch chan<- mdl.SendMessageResponse)
	// GetMessage gets a specific message of a specific user
	GetMessage(ctx context.Context, req *svc.GetMessageReq, ch chan<- mdl.GetMessageResponse)
	// GetMessages gets all messages of a specific user
	GetMessages(ctx context.Context, req *common.MailBoxID, ch chan<- mdl.GetMessageResponse)
	// GetSentMessages gets all the messages sent by a specific user
	GetSentMessages(ctx context.Context, req *common.MailBoxID, ch chan<- mdl.GetMessageResponse)
	// GetReceivedMessages gets all the messages received by a specific user
	GetReceivedMessages(ctx context.Context, req *common.MailBoxID, ch chan<- mdl.GetMessageResponse)
	// GetMessageByParent gets all the replies hierarchy of a specific message for a specific user
	GetMessagesByParent(ctx context.Context, req *svc.GetMessagesByParentReq, ch chan<- mdl.GetMessageResponse)
	// MarkAsRead marks a message as read
	MarkAsRead(ctx context.Context, req *svc.GetMessageReq, ch chan<- error)
	// DeleteMessage deletes a message from messages table
	DeleteMessage(ctx context.Context, req *svc.DeleteMessageReq, ch chan<- error)
	// DeleteMessage deletes a list of messages from messages table
	DeleteMessages(ctx context.Context, req *svc.DeleteMessagesReq, ch chan<- error)
}

func CreateMessagesHandler(dbConf *config.Messages, logger *logrus.Logger) (MessagesHandler, error) {
	var messagesHandler MessagesHandler
	switch dbConf.Type {
	case "dynamodb":
		messagesHandler = dynamo.NewMessagesRepo(dbConf, logger)
	default:
		return nil, fmt.Errorf("%s is an unknown database type", dbConf.Type)
	}
	return messagesHandler, nil
}
