package utils

import (
	"github.com/bxcodec/faker/v3"
	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
)

// TTIsNilOrEmpty represents a testcase for IsNilOrEmpty function
type TTIsNilOrEmpty struct {
	Name         string
	Object       interface{}
	IsNilOrEmpty bool
}

// CreateTTIsNilOrEmpty generates different test cases for IsNilOrEmpty function
func CreateTTIsNilOrEmpty() []TTIsNilOrEmpty {
	// create nil objects
	var nilUserID *common.UserID
	var nilMessageID *common.MessageID
	var nilMailboxID *common.MailBoxID
	var nilGetMessageReq *svc.GetMessageReq
	var nilGetMessageByParentReq *svc.GetMessagesByParentReq
	return []TTIsNilOrEmpty{
		{
			Name:         "valid userID",
			Object:       &common.UserID{Value: faker.UUIDHyphenated()},
			IsNilOrEmpty: false,
		},
		{
			Name:         "empty userID",
			Object:       &common.UserID{Value: ""},
			IsNilOrEmpty: true,
		},
		{
			Name:         "nil userID",
			Object:       nilUserID,
			IsNilOrEmpty: true,
		},
		{
			Name:         "valid MessageID",
			Object:       &common.MessageID{Value: faker.UUIDHyphenated()},
			IsNilOrEmpty: false,
		},
		{
			Name:         "empty MessageID",
			Object:       &common.MessageID{Value: ""},
			IsNilOrEmpty: true,
		},
		{
			Name:         "nil MessageID",
			Object:       nilMessageID,
			IsNilOrEmpty: true,
		},
		{
			Name:         "valid MailboxID",
			Object:       &common.MailBoxID{Value: faker.UUIDHyphenated()},
			IsNilOrEmpty: false,
		},
		{
			Name:         "empty MailboxID",
			Object:       &common.MailBoxID{Value: ""},
			IsNilOrEmpty: true,
		},
		{
			Name:         "nil MailboxID",
			Object:       nilMailboxID,
			IsNilOrEmpty: true,
		},
		{
			Name:         "valid getMessageReq",
			Object:       &svc.GetMessageReq{MessageID: &common.MessageID{Value: "not nil"}, MailboxID: &common.MailBoxID{Value: "not nil"}},
			IsNilOrEmpty: false,
		},
		{
			Name:         "nil getMessageReq",
			Object:       nilGetMessageReq,
			IsNilOrEmpty: true,
		},
		{
			Name:         "empty getMessageReq: empty messageID",
			Object:       &svc.GetMessageReq{MessageID: nilMessageID, MailboxID: &common.MailBoxID{Value: "not nil"}},
			IsNilOrEmpty: true,
		},
		{
			Name:         "empty getMessageReq: empty mailboxID",
			Object:       &svc.GetMessageReq{MailboxID: nilMailboxID, MessageID: &common.MessageID{Value: "not nil"}},
			IsNilOrEmpty: true,
		},
		{
			Name:         "valid GetMessagesByParentReq",
			Object:       &svc.GetMessagesByParentReq{ParentMessageID: &common.MessageID{Value: "not nil"}, MailboxID: &common.MailBoxID{Value: "not nil"}},
			IsNilOrEmpty: false,
		},
		{
			Name:         "nil GetMessagesByParentReq",
			Object:       nilGetMessageByParentReq,
			IsNilOrEmpty: true,
		},
		{
			Name:         "empty GetMessagesByParentReq: empty ParentMessage",
			Object:       &svc.GetMessagesByParentReq{ParentMessageID: nilMessageID, MailboxID: &common.MailBoxID{Value: "not nil"}},
			IsNilOrEmpty: true,
		},
		{
			Name:         "empty GetMessagesByParentReq: empty mailboxID",
			Object:       &svc.GetMessagesByParentReq{MailboxID: nilMailboxID, ParentMessageID: &common.MessageID{Value: "not nil"}},
			IsNilOrEmpty: true,
		},
		{
			Name:         "nil input",
			Object:       nil,
			IsNilOrEmpty: true,
		},
		{
			Name:         "another type",
			Object:       struct{ NoneSense string }{},
			IsNilOrEmpty: false,
		},
	}
}
