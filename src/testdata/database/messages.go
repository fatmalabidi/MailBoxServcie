package database

import (
	"math/rand"

	"github.com/bxcodec/faker/v3"
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
)

// TTSendMessage represents a test case for SendMessage function
type TTSendMessage struct {
	Name     string
	Conf     *config.Messages
	Req      *svc.SendMessageReq
	HasError bool
}

// CreateTTsSendMessage generates different testcases for SendMessage function
func CreateTTsSendMessage(conf *config.Messages) []TTSendMessage {
	// make valid request
	validReq := getValidSendRequest()
	// make invalid request with missing mandatory field
	inValidReq := &svc.SendMessageReq{}
	_ = faker.FakeData(inValidReq)
	inValidReq.MailboxID = &common.MailBoxID{
		Value: "",
	}
	// make invalid config
	invalidConf := *conf
	invalidConf.TableName = "unknown_name"

	return []TTSendMessage{
		{
			Name:     "valid request",
			Conf:     conf,
			Req:      validReq,
			HasError: false,
		},
		{
			Name:     "invalid config",
			Conf:     &invalidConf,
			Req:      validReq,
			HasError: true,
		},
		{
			Name:     "missing mandatory field in Request",
			Conf:     conf,
			Req:      inValidReq,
			HasError: true,
		},
	}

}

// CreateTTsReceiveMessage generates different testcases for ReceiveMessage function
func CreateTTsReceiveMessage(conf *config.Messages) []TTSendMessage {
	return CreateTTsSendMessage(conf)
}

// getValidSendRequest returns a valid request for adding a message to db
func getValidSendRequest() *svc.SendMessageReq {
	validReq := svc.SendMessageReq{}
	_ = faker.FakeData(&validReq)
	validReq.MailboxID = &common.MailBoxID{
		Value: faker.UUIDHyphenated(),
	}
	return &validReq
}

// TTGetMessage represents a test case for GetMessage
type TTGetMessage struct {
	Name     string
	Conf     *config.Messages
	GetReq   func(messageID string) *svc.GetMessageReq
	HasError bool
}

// CreateTTsGetMessage generates different testcases for GetMessage function
func CreateTTsGetMessage(conf *config.Messages) ([]TTGetMessage, *svc.SendMessageReq) {
	// create a request to prepare a message to get
	addedMsg := getValidSendRequest()
	// create a function that generates valid get request
	getValidReq := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: addedMsg.MailboxID,
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates invalid get request : missing fields
	getInvalidRequestNoID := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: &common.MailBoxID{Value: ""},
			MessageID: &common.MessageID{Value: ""},
		}
	}
	// create a function that generates invalid get request : unexisting message
	getInvalidRequestWrongID := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: &common.MailBoxID{Value: faker.UUIDHyphenated()},
			MessageID: &common.MessageID{Value: faker.UUIDHyphenated()},
		}
	}
	// make invalid config
	invalidConf := *conf
	invalidConf.TableName = "unknown_name"

	data := []TTGetMessage{
		{
			Name:     "valid request",
			Conf:     conf,
			GetReq:   getValidReq,
			HasError: false,
		},
		{
			Name:     "invalid config",
			Conf:     &invalidConf,
			GetReq:   getValidReq,
			HasError: true,
		},
		{
			Name:     "missing mandatory field in Request",
			Conf:     conf,
			GetReq:   getInvalidRequestNoID,
			HasError: true,
		},
		{
			Name:     "unexisting message",
			Conf:     conf,
			GetReq:   getInvalidRequestWrongID,
			HasError: true,
		},
	}
	return data, addedMsg
}

// TTDeleteMessage represents a test case for DeleteMessage functions
type TTDeleteMessage struct {
	Name     string
	Conf     *config.Messages
	GetReq   func(messageID string) *svc.DeleteMessageReq
	HasError bool
}

// addChildrenRequest returns list of sendMessageReq with parentID entered as parameter
type addChildrenRequests func(parentID string, childrenNumber int) []*svc.SendMessageReq

// CreateTTsDeleteMessage generates different testcases for DeleteMessage function
// it returns also a SendRequest to insert parent message in db,
// and a function that generates SendRequest for children messages
func CreateTTsDeleteMessage(conf *config.Messages) ([]TTDeleteMessage, *svc.SendMessageReq, addChildrenRequests) {
	// create a request to prepare a parent message to delete
	parentToAdd := getValidSendRequest()
	// create a function that generates SendRequests to prepare children message to delete
	childrenToAdd := func(parentID string, childrenNumber int) []*svc.SendMessageReq {
		var childrenRequests []*svc.SendMessageReq
		for i := 0; i < childrenNumber; i++ {
			req := getValidSendRequest()
			req.MailboxID = parentToAdd.MailboxID
			req.ParentMessageID.Value = parentID
			childrenRequests = append(childrenRequests, req)
		}
		return childrenRequests
	}
	// create a function that generates valid DeleteRequest for deletion: without children
	getValidReqWithoutChildren := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID: parentToAdd.MailboxID,
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates valid DeleteRequest for deletion: with children
	getValidReqWithChildren := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID:       parentToAdd.MailboxID,
			MessageID:       &common.MessageID{Value: messageID},
			WithDescendants: true,
		}
	}
	// create a function that generates invalid DeleteRequest : missing mailboxID
	getInvalidRequestNoMBID := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID:       &common.MailBoxID{Value: ""},
			MessageID:       &common.MessageID{Value: messageID},
			WithDescendants: false,
		}
	}
	// create a function that generates invalid DeleteRequest : missing messageID
	getInvalidRequestNoMsgID := func(messageID string) *svc.DeleteMessageReq {
		return &svc.DeleteMessageReq{
			MailboxID: parentToAdd.MailboxID,
			MessageID: &common.MessageID{Value: ""},
		}
	}
	// make invalid config
	invalidConf := *conf
	invalidConf.TableName = "unknown_name"

	data := []TTDeleteMessage{
		{
			Name:     "invalid config",
			Conf:     &invalidConf,
			GetReq:   getValidReqWithoutChildren,
			HasError: true,
		},
		{
			Name:     "missing messageID in Request",
			Conf:     conf,
			GetReq:   getInvalidRequestNoMsgID,
			HasError: true,
		},
		{
			Name:     "missing mailboxID in Request",
			Conf:     conf,
			GetReq:   getInvalidRequestNoMBID,
			HasError: true,
		},
		{
			Name:     "valid request: message without children",
			Conf:     conf,
			GetReq:   getValidReqWithoutChildren,
			HasError: false,
		},
		{
			Name:     "valid request: message with children",
			Conf:     conf,
			GetReq:   getValidReqWithChildren,
			HasError: false,
		},
	}
	return data, parentToAdd, childrenToAdd
}

// TTDeleteMessages represents a test case for DeleteMessages function
type TTDeleteMessages struct {
	Name     string
	Conf     *config.Messages
	GetReq   func(messageIDs []*common.MessageID) *svc.DeleteMessagesReq
	HasError bool
}

// CreateTTsDeleteMessages generates different testcases for DeleteMessages function
func CreateTTsDeleteMessages(conf *config.Messages) ([]TTDeleteMessages, []*svc.SendMessageReq) {
	// create requests to prepare a list of messages to delete
	sendRequests, _, mailboxID := createMessagesReqs(15, 0)
	// create a function that generates valid deleteMessages request for deletion
	getValidReq := func(messageIDs []*common.MessageID) *svc.DeleteMessagesReq {
		return &svc.DeleteMessagesReq{
			MailboxID:  mailboxID,
			MessageIDs: messageIDs,
		}
	}
	// create a function that generates invalid deleteMessages request : missing field
	getInvalidRequestNoID := func(messageIDs []*common.MessageID) *svc.DeleteMessagesReq {
		return &svc.DeleteMessagesReq{
			MailboxID:  &common.MailBoxID{Value: ""},
			MessageIDs: messageIDs,
		}
	}
	// make invalid config
	invalidConf := *conf
	invalidConf.TableName = "unknown_name"

	data := []TTDeleteMessages{
		{
			Name:     "invalid config",
			Conf:     &invalidConf,
			GetReq:   getValidReq,
			HasError: true,
		},
		{
			Name:     "missing mandatory field in Request",
			Conf:     conf,
			GetReq:   getInvalidRequestNoID,
			HasError: true,
		},
		{
			Name:     "valid request",
			Conf:     conf,
			GetReq:   getValidReq,
			HasError: false,
		},
	}
	return data, sendRequests
}

// TTGetMessages represents a test case for GetMessages function
type TTGetMessages struct {
	Name                   string
	Conf                   *config.Messages
	MailboxID              *common.MailBoxID
	ExpectedSentResult     int
	ExpectedReceivedResult int
	HasError               bool
}

// CreateTTsGetMessages generates different testcases for GetMessages function
func CreateTTsGetMessages(conf *config.Messages) ([]TTGetMessages, []*svc.SendMessageReq, []*svc.SendMessageReq) {
	// decide the number of messages to send and to receive
	nSend := rand.Intn(15) + 1
	nReceive := rand.Intn(15) + 1
	// create a request to prepare send and receive messages
	sendMsgsReq, receiveMsgsReq, mailboxID := createMessagesReqs(nSend, nReceive)
	// make mailbox with empty messages
	emptyMailbox := &common.MailBoxID{Value: faker.UUIDHyphenated()}
	// make invalid mailbox
	invalidMailbox := &common.MailBoxID{Value: ""}
	// make invalid config
	invalidConf := *conf
	invalidConf.TableName = "unknown_name"
	data := []TTGetMessages{
		{
			Name:                   "valid request: existing messages",
			Conf:                   conf,
			MailboxID:              mailboxID,
			ExpectedSentResult:     nSend,
			ExpectedReceivedResult: nReceive,
			HasError:               false,
		},
		{
			Name:      "valid request: empty messages",
			Conf:      conf,
			MailboxID: emptyMailbox,
			HasError:  false,
		},
		{
			Name:      "invalid mailbox",
			Conf:      conf,
			MailboxID: invalidMailbox,
			HasError:  true,
		},
		{
			Name:      "invalid config",
			Conf:      &invalidConf,
			MailboxID: mailboxID,
			HasError:  true,
		},
	}
	return data, sendMsgsReq, receiveMsgsReq
}

// createMessagesReqs generates a list of SendMessage and ReceiveMessage requests and returns them with the Mailbox they belong to
func createMessagesReqs(numberSend int, numberReceive int) ([]*svc.SendMessageReq, []*svc.SendMessageReq, *common.MailBoxID) {
	mailboxID := &common.MailBoxID{Value: faker.UUIDHyphenated()}
	var sendRequests []*svc.SendMessageReq
	var receiveRequests []*svc.SendMessageReq
	// prepare send requests
	for i := 0; i < numberSend; i++ {
		validReq := svc.SendMessageReq{}
		_ = faker.FakeData(&validReq)
		validReq.MailboxID = mailboxID
		sendRequests = append(sendRequests, &validReq)
	}
	// prepare receive requests
	for i := 0; i < numberReceive; i++ {
		validReq := svc.SendMessageReq{}
		_ = faker.FakeData(&validReq)
		validReq.MailboxID = mailboxID
		receiveRequests = append(receiveRequests, &validReq)
	}
	return sendRequests, receiveRequests, mailboxID
}

// CreateTTsGetSentMessages generates different testcases for GetSentMessages function
func CreateTTsGetSentMessages(conf *config.Messages) ([]TTGetMessages, []*svc.SendMessageReq, []*svc.SendMessageReq) {
	return CreateTTsGetMessages(conf)
}

// CreateTTsGetReceivedMessages generates different testcases for GetReceivedMessages function
func CreateTTsGetReceivedMessages(conf *config.Messages) ([]TTGetMessages, []*svc.SendMessageReq, []*svc.SendMessageReq) {
	return CreateTTsGetMessages(conf)
}

// TTGetMessagesByParent represents a test case for GetMessagesByParent function
type TTGetMessagesByParent struct {
	Name           string
	Conf           *config.Messages
	Req            *svc.GetMessagesByParentReq
	ExpectedNumber int
	HasError       bool
}

// CreateTTsGetMessagesByParent generates different testcases for GetMessagesByParent function
func CreateTTsGetMessagesByParent(conf *config.Messages) ([]TTGetMessagesByParent, []*svc.SendMessageReq, []*svc.SendMessageReq) {
	// decide the number of messages to send and to receive
	nSend := rand.Intn(10) + 1
	nReceive := rand.Intn(10) + 1
	// prepare the parentMessageID of the messages to create
	parentMessageID := faker.UUIDHyphenated()
	// create a request to prepare send and receive messages
	sendMsgsReq, receiveMsgsReq, mailboxID := createMessagesReqs(nSend, nReceive)
	// set the parentMessageID to all messages to create
	setParentID(parentMessageID, sendMsgsReq)
	setParentID(parentMessageID, receiveMsgsReq)
	// make invalid config
	invalidConf := *conf
	invalidConf.TableName = "unknown_name"
	// create a valid request : existing children messages
	validRequestEC := &svc.GetMessagesByParentReq{
		MailboxID:       mailboxID,
		ParentMessageID: &common.MessageID{Value: parentMessageID},
	}
	// create a valid request : non existing children messages
	validRequestNEC := &svc.GetMessagesByParentReq{
		MailboxID:       mailboxID,
		ParentMessageID: &common.MessageID{Value: faker.UUIDHyphenated()},
	}
	// create an invalid request: non existing mailbox
	invalidRequestNEMB := &svc.GetMessagesByParentReq{
		MailboxID:       &common.MailBoxID{Value: faker.UUIDHyphenated()},
		ParentMessageID: &common.MessageID{Value: parentMessageID},
	}

	// create an invalid request: missing mandatory fields
	invalidRequestMMF := &svc.GetMessagesByParentReq{
		MailboxID:       &common.MailBoxID{Value: ""},
		ParentMessageID: &common.MessageID{Value: parentMessageID},
	}

	data := []TTGetMessagesByParent{
		{
			Name:           "Valid Request: existing children",
			Conf:           conf,
			Req:            validRequestEC,
			ExpectedNumber: nReceive + nSend,
			HasError:       false,
		},
		{
			Name:           "Valid Request: non existing children",
			Conf:           conf,
			Req:            validRequestNEC,
			ExpectedNumber: 0,
			HasError:       false,
		},
		{
			Name:     "invalid config",
			Conf:     &invalidConf,
			Req:      validRequestEC,
			HasError: true,
		},
		{
			Name:     "non existing mailbox",
			Conf:     conf,
			Req:      invalidRequestNEMB,
			HasError: true,
		},
		{
			Name:     "missing mandatory field",
			Conf:     conf,
			Req:      invalidRequestMMF,
			HasError: true,
		},
	}
	return data, sendMsgsReq, receiveMsgsReq
}

func setParentID(parentMessageID string, requests []*svc.SendMessageReq) {
	for _, request := range requests {
		request.ParentMessageID.Value = parentMessageID
	}
}

// TTMarkAsRead represents a test case for MarkAsRead function
type TTMarkAsRead struct {
	Name     string
	Conf     *config.Messages
	GetReq   func(string) *svc.GetMessageReq
	HasError bool
}

// CreateTTsMarkAsRead generates different test cases for MarkAsRead function
func CreateTTsMarkAsRead(conf *config.Messages) ([]TTMarkAsRead, *svc.SendMessageReq) {
	// create a request to prepare a message to mark as read
	addedMsg := getValidSendRequest()
	// create a function that generates valid get request for deletion
	getValidReq := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: addedMsg.MailboxID,
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates invalid get request : missing mailboxID
	getInvalidRequestNoMBID := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: &common.MailBoxID{Value: ""},
			MessageID: &common.MessageID{Value: messageID},
		}
	}
	// create a function that generates invalid get request : missing MessageID
	getInvalidRequestNoMID := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: addedMsg.MailboxID,
			MessageID: &common.MessageID{Value: ""},
		}
	}
	// create a function that generates invalid get request : non existing message
	getInvalidRequestNotExist := func(messageID string) *svc.GetMessageReq {
		return &svc.GetMessageReq{
			MailboxID: &common.MailBoxID{Value: faker.UUIDHyphenated()},
			MessageID: &common.MessageID{Value: faker.UUIDHyphenated()},
		}
	}

	// make invalid config
	invalidConf := *conf
	invalidConf.TableName = "unknown_name"

	data := []TTMarkAsRead{
		{
			Name:     "valid request",
			Conf:     conf,
			GetReq:   getValidReq,
			HasError: false,
		},
		{
			Name:     "invalid config",
			Conf:     &invalidConf,
			GetReq:   getValidReq,
			HasError: true,
		},
		{
			Name:     "missing mandatory field in Request: mailboxID",
			Conf:     conf,
			GetReq:   getInvalidRequestNoMBID,
			HasError: true,
		},
		{
			Name:     "missing mandatory field in Request: MessageID",
			Conf:     conf,
			GetReq:   getInvalidRequestNoMID,
			HasError: true,
		},
		{
			Name:     "non existing message",
			Conf:     conf,
			GetReq:   getInvalidRequestNotExist,
			HasError: true,
		},
	}
	return data, addedMsg
}
