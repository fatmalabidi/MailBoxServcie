package service

import (
	"fmt"
	"math/rand"

	"github.com/bxcodec/faker/v3"
	"github.com/fatmalabidi/protobuf/common"
	"github.com/fatmalabidi/protobuf/mailbox"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/google/uuid"
)

type TSendMessage struct {
	Name     string
	HasError bool
	Req      *svc.SendMessageReq
}

type TGetMessages struct {
	Name             string
	HasError         bool
	Id               *common.MailBoxID
	ExpectedMessages int
}

type TMarkAsRead struct {
	Name     string
	HasError bool
	Req      *svc.GetMessageReq
}

type TGetReceivedMessages struct {
	Name             string
	HasError         bool
	Id               *common.MailBoxID
	ExpectedMessages int
}

// CreateTTSendMessage prepares test data for SendMessage service
func CreateTTSendMessage(parenIDNoReplay, parentIDWithReplay string, mailBoxID *common.MailBoxID) []TSendMessage {
	senderID := &common.UserID{Value: uuid.New().String()}
	receiverID := &common.UserID{Value: uuid.New().String()}
	return []TSendMessage{
		// the first testcase will be used also for invalid context and invalid config
		{
			Name:     "valid data",
			HasError: false,
			Req:      createSendMailReq(senderID, receiverID, false, nil, mailBoxID),
		},
		{
			Name:     "nil mailboxID",
			HasError: true,
			Req:      createSendMailReq(senderID, receiverID, false, nil, nil),
		},
		{
			Name:     "invalid parentID format",
			HasError: true,
			Req:      createSendMailReq(senderID, receiverID, false, &common.MessageID{Value: "invalid parent ID"}, mailBoxID),
		},
		{
			Name:     "nil senderID",
			HasError: true,
			Req:      createSendMailReq(nil, receiverID, false, nil, mailBoxID),
		},
		{
			Name:     "invalid senderID",
			HasError: true,
			Req:      createSendMailReq(&common.UserID{Value: "invalid uuid"}, receiverID, false, nil, mailBoxID),
		},
		{
			Name:     "nil receiverID",
			HasError: true,
			Req:      createSendMailReq(senderID, nil, false, nil, mailBoxID),
		},
		{
			Name:     "invalid receiver",
			HasError: true,
			Req:      createSendMailReq(senderID, &common.UserID{Value: "invalid uuid"}, false, nil, mailBoxID),
		},

		{
			Name:     "message with parent-id: no-replay = false",
			HasError: false,
			Req:      createSendMailReq(senderID, receiverID, false, &common.MessageID{Value: parentIDWithReplay}, mailBoxID),
		},
		{
			Name:     "message with parent-id: no-replay = true",
			HasError: true,
			Req:      createSendMailReq(senderID, receiverID, true, &common.MessageID{Value: parenIDNoReplay}, mailBoxID),
		},
	}
}

// createSendMailReq creates a message info
// it is public because it will be used later in the
// test to test the case of sending a message for a parent with no-reply=true
func createSendMailReq(senderID, receiverID *common.UserID, noReply bool, ParentMessageID *common.MessageID, MailboxID *common.MailBoxID) *svc.SendMessageReq {
	return &svc.SendMessageReq{
		MessageInfo: &mailbox.MailMessageInfo{
			SenderID:        senderID,
			RecipientID:     receiverID,
			SenderFirstName: "test sender name",
			SenderLastName:  "test sender last name ",
			Title:           "test message title",
			MessageBody:     "test message body",
			MessageType:     common.MessageType_INFO,
			NoReply:         noReply,
		},
		ParentMessageID: ParentMessageID,
		MailboxID:       MailboxID,
	}
}

// CreateParentReq
func CreateParentReq(mailBoxID *common.MailBoxID) (*svc.SendMessageReq, *svc.SendMessageReq) {
	// sendReqNoReply is with no-reply = false
	sendReqWithNoReply := createSendMailReq(&common.UserID{Value: uuid.New().String()}, &common.UserID{Value: uuid.New().String()}, true, nil, mailBoxID)
	// sendReqNoReply is with no-reply = true
	sendReqWithReply := createSendMailReq(&common.UserID{Value: uuid.New().String()}, &common.UserID{Value: uuid.New().String()}, false, nil, mailBoxID)
	return sendReqWithNoReply, sendReqWithReply
}

// GetRequest represents a function that receives a messageID and returns a getMessageRequest
type GetRequest func(messageID string) *svc.GetMessageReq

// TTGetMessage represent a testcase for GetMessage function
type TTGetMessage struct {
	Name            string
	GetReq          GetRequest
	ContextDeadline int64 // in seconds
	HasError        bool
}

// CreateTTGetMessage generates different test cases for GetMessage function and a request for adding a message
func CreateTTGetMessage() ([]TTGetMessage, *svc.SendMessageReq) {
	// prepare a request to add a message
	messageToAdd, mailboxID := getValidSendRequest()
	// create a function that generates a valid request
	getValidReq := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: mailboxID,
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates an invalid request: nil mailboxID
	getInvalidRequestNilMB := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: nil,
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates an invalid request: invalid mailboxID
	getInvalidRequestInvMB := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: &common.MailBoxID{Value: "invalid"},
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates an invalid request: nil messageID
	getInvalidRequestNilM := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: mailboxID,
			MessageID: nil,
		}
	}
	// create a function that generates an invalid request: invalid messageID
	getInvalidRequestInvM := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: mailboxID,
			MessageID: &common.MessageID{Value: "invalid"},
		}
	}
	// create a function that generates an invalid request: non existing message
	getInvalidRequestNotExist := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: &common.MailBoxID{Value: faker.UUIDHyphenated()},
			MessageID: &common.MessageID{Value: faker.UUIDHyphenated()},
		}
	}

	data := []TTGetMessage{
		{
			Name:            "valid request",
			GetReq:          getValidReq,
			ContextDeadline: 10,
			HasError:        false,
		},
		{
			Name:            "invalid request: nil mailboxID",
			GetReq:          getInvalidRequestNilMB,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: invalid mailboxID",
			GetReq:          getInvalidRequestInvMB,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: nil messageID",
			GetReq:          getInvalidRequestNilM,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: invalid messageID",
			GetReq:          getInvalidRequestInvM,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid context",
			GetReq:          getValidReq,
			ContextDeadline: 100,
			HasError:        true,
		},
		{
			Name:            "invalid request: non existing message",
			GetReq:          getInvalidRequestNotExist,
			ContextDeadline: 10,
			HasError:        true,
		},
	}
	return data, messageToAdd
}

// getValidSendRequest create s a valid senMessage request and returns it with the used mailboxID
func getValidSendRequest() (*svc.SendMessageReq, *common.MailBoxID) {
	mailboxID := &common.MailBoxID{Value: faker.UUIDHyphenated()}
	senderID := &common.UserID{Value: faker.UUIDHyphenated()}
	receiverID := &common.UserID{Value: faker.UUIDHyphenated()}
	parentMessageID := &common.MessageID{Value: faker.UUIDHyphenated()}
	req := createSendMailReq(senderID, receiverID, true, parentMessageID, mailboxID)
	return req, mailboxID
}

func CreateSendMessageReqs(mailboxID *common.MailBoxID, messageCount int) []*svc.SendMessageReq {
	var messagesToSend []*svc.SendMessageReq
	for i := 0; i < messageCount; i++ {
		messagesToSend = append(messagesToSend, &svc.SendMessageReq{
			MessageInfo: &mailbox.MailMessageInfo{
				SenderID:        &common.UserID{Value: uuid.New().String()},
				RecipientID:     &common.UserID{Value: uuid.New().String()},
				SenderFirstName: "test sender name",
				SenderLastName:  "test sender last name ",
				Title:           fmt.Sprintf("test message title -%d-", i+1),
				MessageBody:     "test message body",
				MessageType:     common.MessageType_INFO,
				NoReply:         false,
			},
			MailboxID: mailboxID,
		})
	}
	return messagesToSend
}

func CreateGetMessageReqs(mailboxID *common.MailBoxID, ids []string) []*svc.GetMessageReq {
	var messagesToGet []*svc.GetMessageReq
	for _, id := range ids {
		messagesToGet = append(messagesToGet, &svc.GetMessageReq{
			MailboxID: mailboxID,
			MessageID: &common.MessageID{Value: id},
		})
	}
	return messagesToGet
}

func CreateTTGetMessages(mailboxID *common.MailBoxID, expected int) []TGetMessages {
	return []TGetMessages{
		{
			Name:             "valid request",
			HasError:         false,
			Id:               mailboxID,
			ExpectedMessages: expected,
		},
		{
			Name:     "valid request with unknown mailboxID",
			HasError: false,
			Id:       &common.MailBoxID{Value: uuid.New().String()},
		},
		{
			Name:     "invalid mailboxID",
			HasError: true,
			Id:       &common.MailBoxID{Value: "invalid-ID"},
		},
	}
}

// TTGetSentMessages represents a testcase for GetSentMessages function
type TTGetSentMessages struct {
	Name            string
	MailboxID       *common.MailBoxID
	ExpectedResults int
	ContextDeadline int64 // in seconds
	HasError        bool
}

// CreateTTGetSentMessages generates different test cases for GetSentMessages function
func CreateTTGetSentMessages() ([]TTGetSentMessages, []*svc.SendMessageReq, []*svc.SendMessageReq) {
	// prepare the mailbox to use
	mailboxID := &common.MailBoxID{Value: faker.UUIDHyphenated()}
	// prepare the number of messages to add
	numberSentMessage := rand.Intn(10) + 1
	numberReceivedMessage := rand.Intn(10) + 1
	// prepare sent messages to add to db
	sentMessagesToAdd := CreateSendMessageReqs(mailboxID, numberSentMessage)
	// prepare received messages to add to db
	receivedMessagesToAdd := CreateSendMessageReqs(mailboxID, numberReceivedMessage)
	data := []TTGetSentMessages{
		{
			Name:            "valid request: existing messages",
			MailboxID:       mailboxID,
			ExpectedResults: numberSentMessage,
			HasError:        false,
			ContextDeadline: 10,
		},
		{
			Name:            "valid request: non existing messages",
			MailboxID:       &common.MailBoxID{Value: faker.UUIDHyphenated()},
			ExpectedResults: 0,
			HasError:        false,
			ContextDeadline: 10,
		},
		{
			Name:            "invalid request: invalid mailboxID format",
			MailboxID:       &common.MailBoxID{Value: "invalid-format"},
			ExpectedResults: 0,
			HasError:        true,
			ContextDeadline: 10,
		},
		{
			Name:            "invalid context",
			MailboxID:       mailboxID,
			HasError:        true,
			ContextDeadline: 100,
		},
	}
	return data, sentMessagesToAdd, receivedMessagesToAdd
}

func CreateTTMarkAsRead(mailboxID *common.MailBoxID, messageID *common.MessageID) []TMarkAsRead {
	return []TMarkAsRead{
		{
			Name:     "valid request",
			HasError: false,
			Req: &svc.GetMessageReq{
				MailboxID: mailboxID,
				MessageID: messageID,
			},
		},
		{
			Name:     "unknown mailboxID",
			HasError: true,
			Req: &svc.GetMessageReq{
				MailboxID: &common.MailBoxID{Value: uuid.New().String()},
				MessageID: messageID,
			},
		},
		{
			Name:     "unknown messageID",
			HasError: true,
			Req: &svc.GetMessageReq{
				MailboxID: mailboxID,
				MessageID: &common.MessageID{Value: uuid.New().String()},
			},
		},
		{
			Name:     "invalid mailboxID",
			HasError: true,
			Req: &svc.GetMessageReq{
				MailboxID: &common.MailBoxID{Value: "invalid-ID"},
				MessageID: messageID,
			},
		},
		{
			Name:     "invalid messageID",
			HasError: true,
			Req: &svc.GetMessageReq{
				MailboxID: mailboxID,
				MessageID: &common.MessageID{Value: "invalid-ID"},
			},
		},
		{
			Name:     "empty request",
			HasError: true,
			Req:      &svc.GetMessageReq{},
		},
	}
}
func CreateTTGetReceivedMessages(mailboxID *common.MailBoxID, expected int) []TGetReceivedMessages {
	return []TGetReceivedMessages{
		{
			Name:             "valid request",
			HasError:         false,
			Id:               mailboxID,
			ExpectedMessages: expected,
		},
		{
			Name:     "unknown mailboxID",
			HasError: true,
			Id:       &common.MailBoxID{Value: uuid.New().String()},
		},
		{
			Name:     "invalid mailboxID",
			HasError: true,
			Id:       &common.MailBoxID{Value: "invalid-ID"},
		},
		{
			Name:     "empty mailboxID",
			HasError: true,
			Id:       &common.MailBoxID{Value: ""},
		},
	}
}

// TTGetMessagesByParent represents a test case for GetMessagesByParent function
type TTGetMessagesByParent struct {
	Name             string
	Req              *svc.GetMessagesByParentReq
	ExpectedMessages int
	ContextDeadline  int64 // in seconds
	HasError         bool
}

// CreateTTGetMessagesByParent generates different test cases for GetMessagesByParent function
func CreateTTGetMessagesByParent() ([]TTGetMessagesByParent, []*svc.SendMessageReq, []*svc.SendMessageReq) {
	// decide the number of messages to send and to receive
	nSend := rand.Intn(10) + 1
	nReceive := rand.Intn(10) + 1
	// prepare the parentMessageID of the messages to create
	parentMessageID := &common.MessageID{Value: faker.UUIDHyphenated()}
	// prepare the mailbox to use
	mailboxID := &common.MailBoxID{Value: faker.UUIDHyphenated()}
	// prepare sent messages to add to db
	sentMessagesToAdd := CreateSendMessageReqs(mailboxID, nSend)
	// prepare received messages to add to db
	receivedMessagesToAdd := CreateSendMessageReqs(mailboxID, nReceive)
	// set parentMessageIDID to the messages to add to db
	for _, req := range append(sentMessagesToAdd, receivedMessagesToAdd...) {
		req.ParentMessageID = parentMessageID
	}
	// create a valid request : existing children messages
	validRequestEC := &svc.GetMessagesByParentReq{
		MailboxID:       mailboxID,
		ParentMessageID: parentMessageID,
	}
	// create a valid request : non existing children messages
	validRequestNEC := &svc.GetMessagesByParentReq{
		MailboxID:       mailboxID,
		ParentMessageID: &common.MessageID{Value: faker.UUIDHyphenated()},
	}
	// create aninvalid request: non existing mailbox
	invalidRequestNEMB := &svc.GetMessagesByParentReq{
		MailboxID:       &common.MailBoxID{Value: faker.UUIDHyphenated()},
		ParentMessageID: parentMessageID,
	}
	// create an invalid request: nil mailboxID
	invalidRequestNilMB := &svc.GetMessagesByParentReq{
		MailboxID:       nil,
		ParentMessageID: parentMessageID,
	}
	// create an invalid request: empty mailboxID
	invalidRequestEMB := &svc.GetMessagesByParentReq{
		MailboxID:       &common.MailBoxID{Value: ""},
		ParentMessageID: parentMessageID,
	}
	// create an invalid request: nil parentMessage
	invalidRequestNilPM := &svc.GetMessagesByParentReq{
		MailboxID:       mailboxID,
		ParentMessageID: nil,
	}
	// create an invalid request: empty parentMessage
	invalidRequestEPM := &svc.GetMessagesByParentReq{
		MailboxID:       mailboxID,
		ParentMessageID: &common.MessageID{Value: ""},
	}

	data := []TTGetMessagesByParent{
		{
			Name:             "Valid Request: existing children",
			Req:              validRequestEC,
			ExpectedMessages: nReceive + nSend,
			HasError:         false,
			ContextDeadline:  10,
		},
		{
			Name:             "Valid Request: non existing children",
			Req:              validRequestNEC,
			ExpectedMessages: 0,
			HasError:         false,
			ContextDeadline:  10,
		},
		{
			Name:            "non existing mailbox",
			Req:             invalidRequestNEMB,
			HasError:        true,
			ContextDeadline: 10,
		},
		{
			Name:            "invalid request: nil mailboxID",
			Req:             invalidRequestNilMB,
			HasError:        true,
			ContextDeadline: 10,
		},
		{
			Name:            "invalid request: empty mailboxID",
			Req:             invalidRequestEMB,
			HasError:        true,
			ContextDeadline: 10,
		},
		{
			Name:            "invalid request: nil parentMessage",
			Req:             invalidRequestNilPM,
			HasError:        true,
			ContextDeadline: 10,
		},
		{
			Name:            "invalid request: empty parentMessage",
			Req:             invalidRequestEPM,
			HasError:        true,
			ContextDeadline: 10,
		},
		{
			Name:            "invalid context",
			Req:             validRequestEC,
			HasError:        true,
			ContextDeadline: 100,
		},
	}
	return data, sentMessagesToAdd, receivedMessagesToAdd
}

// TTDeleteMessage represent a testcase for DeleteMessage function
type TTDeleteMessage struct {
	Name            string
	GetReq          func(messageID string) *svc.DeleteMessageReq
	ContextDeadline int64 // in seconds
	HasError        bool
}

// addChildrenRequest returns list of sendMessageReq with parentID entered as parameter
type addChildrenRequests func(parentID string, childrenNumber int) []*svc.SendMessageReq

// CreateTTDeleteMessage generates different test cases for DeleteMessage function
func CreateTTDeleteMessage() ([]TTDeleteMessage, *svc.SendMessageReq, addChildrenRequests) {
	// create a request to prepare a parent message to delete
	parentToAdd, mailboxID := getValidSendRequest()
	// create a function that generates SendRequests to prepare children message to delete
	childrenToAdd := func(parentID string, childrenNumber int) []*svc.SendMessageReq {
		var childrenRequests []*svc.SendMessageReq
		for i := 0; i < childrenNumber; i++ {
			req, _ := getValidSendRequest()
			req.MailboxID = mailboxID
			req.ParentMessageID.Value = parentID
			childrenRequests = append(childrenRequests, req)
		}
		return childrenRequests
	}
	// create a function that generates a valid DeleteMessageRequest with descendants
	getValidReqWithDes := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID:       mailboxID,
			MessageID:       &common.MessageID{Value: messageID},
			WithDescendants: true,
		}
	}
	// create a function that generates a valid DeleteMessageRequest without descendants
	getValidReqWithoutDes := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID:       mailboxID,
			MessageID:       &common.MessageID{Value: messageID},
			WithDescendants: false,
		}
	}
	// create a function that generates an invalid request: nil mailboxID
	getInvalidRequestNilMB := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID: nil,
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates an invalid request: invalid mailboxID
	getInvalidRequestInvMB := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID: &common.MailBoxID{Value: "invalid"},
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates an invalid request: nil messageID
	getInvalidRequestNilM := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID: mailboxID,
			MessageID: nil,
		}
	}
	// create a function that generates an invalid request: invalid messageID
	getInvalidRequestInvM := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID: mailboxID,
			MessageID: &common.MessageID{Value: "invalid"},
		}
	}
	data := []TTDeleteMessage{
		{
			Name:            "invalid request: nil mailbox",
			GetReq:          getInvalidRequestNilMB,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: empty mailbox",
			GetReq:          getInvalidRequestInvMB,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: nil MessageID",
			GetReq:          getInvalidRequestNilM,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: empty MessageID",
			GetReq:          getInvalidRequestInvM,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid context",
			GetReq:          getValidReqWithoutDes,
			ContextDeadline: 100,
			HasError:        true,
		},
		{
			Name:            "valid request: without descendants ",
			GetReq:          getValidReqWithoutDes,
			ContextDeadline: 10,
			HasError:        false,
		},
		{
			Name:            "valid request: with descendants ",
			GetReq:          getValidReqWithDes,
			ContextDeadline: 10,
			HasError:        false,
		},
	}
	return data, parentToAdd, childrenToAdd
}

// TTDeleteMessages represent a testcase for DeleteMessages function
type TTDeleteMessages struct {
	Name            string
	GetReq          func(messageIDs []string) *svc.DeleteMessagesReq
	ContextDeadline int64 // in seconds
	HasError        bool
}

// CreateTTDeleteMessages generates different test cases for DeleteMessages function and a request for adding a message
func CreateTTDeleteMessages() ([]TTDeleteMessages, []*svc.SendMessageReq) {
	// prepare the mailbox to use
	mailboxID := &common.MailBoxID{Value: faker.UUIDHyphenated()}
	// create requests to prepare a list of messages to delete
	numberOfMessage := rand.Intn(5) + 5
	messagesToAdd := CreateSendMessageReqs(mailboxID, numberOfMessage)
	// create a function that converts list of id from string to common.MesageID
	convertToMessageId := func(ids []string) []*common.MessageID {
		var messageIDs []*common.MessageID
		for _, id := range ids {
			messageIDs = append(messageIDs, &common.MessageID{Value: id})
		}
		return messageIDs
	}
	// create a function that generates valid deleteMessages request
	getValidReq := func(ids []string) *svc.DeleteMessagesReq {
		return &svc.DeleteMessagesReq{
			MailboxID:  mailboxID,
			MessageIDs: convertToMessageId(ids),
		}
	}
	// create a function that generates invalid deleteMessages request : nil mailboxID
	getInvalidReqNilMB := func(ids []string) *svc.DeleteMessagesReq {
		return &svc.DeleteMessagesReq{
			MailboxID:  nil,
			MessageIDs: convertToMessageId(ids),
		}
	}
	// create a function that generates invalid deleteMessages request : invalid mailboxID
	getInvalidReqInvMB := func(ids []string) *svc.DeleteMessagesReq {
		return &svc.DeleteMessagesReq{
			MailboxID:  &common.MailBoxID{Value: "invalid"},
			MessageIDs: convertToMessageId(ids),
		}
	}
	// create a function that generates invalid deleteMessages request : nil MessageID
	getInvalidReqNilMsg := func(ids []string) *svc.DeleteMessagesReq {
		return &svc.DeleteMessagesReq{
			MailboxID:  mailboxID,
			MessageIDs: []*common.MessageID{nil},
		}
	}
	// create a function that generates invalid deleteMessages request : invalid MessageID
	getInvalidReqInvMsg := func(ids []string) *svc.DeleteMessagesReq {
		return &svc.DeleteMessagesReq{
			MailboxID:  mailboxID,
			MessageIDs: convertToMessageId([]string{"invalid"}),
		}
	}
	data := []TTDeleteMessages{
		{
			Name:            "invalid request: nil mailbox",
			GetReq:          getInvalidReqNilMB,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: invalid mailbox",
			GetReq:          getInvalidReqInvMB,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: nil MessageID",
			GetReq:          getInvalidReqNilMsg,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid request: invalid MessageID",
			GetReq:          getInvalidReqInvMsg,
			ContextDeadline: 10,
			HasError:        true,
		},
		{
			Name:            "invalid context",
			GetReq:          getValidReq,
			ContextDeadline: 100,
			HasError:        true,
		},
		{
			Name:            "valid request",
			GetReq:          getValidReq,
			ContextDeadline: 10,
			HasError:        false,
		},
	}
	return data, messagesToAdd
}
