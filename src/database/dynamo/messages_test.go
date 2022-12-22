package dynamo_test

import (
	"context"
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/database/dynamo"
	mdl "github.com/fatmalabidi/MailBoxServcie/src/models"
	td "github.com/fatmalabidi/MailBoxServcie/src/testdata/database"
	hlp "github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestRepo_SendMessage(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	//clean the db after the test
	var deleteMessagesReq = &svc.DeleteMessagesReq{}
	defer func() {
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	//create test cases
	tests := td.CreateTTsSendMessage(&conf.Database.Messages)
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the SendMessage request
			ch := make(chan mdl.SendMessageResponse, 1)
			repo.SendMessage(ctx, testCase.Req, ch)
			// get back the result
			res := <-ch
			// check result
			if res.Err == nil && testCase.HasError {
				t.Errorf("expected failure got: %v", res)
			}
			if res.Err != nil && !testCase.HasError {
				t.Errorf("expected success got: %v", res)
			}
			// populate the delete request in case of success adding to db
			if res.MessageID != "" {
				deleteMessagesReq = getDeleteMessagesRequest(testCase.Req.MailboxID, []string{res.MessageID})
			}
		})
	}
}

func TestRepo_ReceiveMessage(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	// clean the db after the test
	var deleteMessagesReq = &svc.DeleteMessagesReq{}
	defer func() {
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// create test cases
	tests := td.CreateTTsReceiveMessage(&conf.Database.Messages)
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the SendMessage request
			ch := make(chan mdl.SendMessageResponse, 1)
			repo.ReceiveMessage(ctx, testCase.Req, ch)
			// get back the result
			res := <-ch
			// check result
			if res.Err == nil && testCase.HasError {
				t.Errorf("expected failure got: %v", res)
			}
			if res.Err != nil && !testCase.HasError {
				t.Errorf("expected success got: %v", res)
			}
			// populate the delete request in case of success adding to db
			if res.MessageID != "" {
				deleteMessagesReq = getDeleteMessagesRequest(testCase.Req.MailboxID, []string{res.MessageID})
			}
		})
	}
}

func TestRepo_GetMessage(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	// create test cases
	tests, addRequest := td.CreateTTsGetMessage(&conf.Database.Messages)
	// insert a message to test getting it
	addRes := sendMessage(addRequest, &conf.Database.Messages, logger)
	if addRes.Err != nil {
		t.Fatalf("unable to insert message for deleteMessage test: %v", addRes.Err)
	}
	// clean the db after the test
	defer func() {
		deleteMessagesReq := getDeleteMessagesRequest(addRequest.MailboxID, []string{addRes.MessageID})
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the getMessage request
			ch := make(chan mdl.GetMessageResponse, 1)
			repo.GetMessage(ctx, testCase.GetReq(addRes.MessageID), ch)
			// get back the result
			getRes := <-ch
			// check result
			if getRes.Err == nil && testCase.HasError {
				t.Errorf("expected failure got: %v", getRes)
			}
			if getRes.Err != nil && !testCase.HasError {
				t.Errorf("expected success got: %v", getRes)
			}
		})
	}
}

func TestRepo_GetMessages(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	// create test cases
	tests, sendRequests, receiveRequests := td.CreateTTsGetMessages(&conf.Database.Messages)
	// insert messages to test getting them
	messageIDs, err := insertMessages(sendRequests, receiveRequests, &conf.Database.Messages, logger)
	if err != nil {
		t.Fatalf("unable to insert message for getMessages test: %v", err)
	}
	// clean the db after the test
	defer func() {
		deleteMessagesReq := &svc.DeleteMessagesReq{
			MailboxID:  sendRequests[0].MailboxID,
			MessageIDs: messageIDs,
		}
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the getMessages request
			ch := make(chan mdl.GetMessageResponse, 1)
			go repo.GetMessages(ctx, testCase.MailboxID, ch)
			// get back the result
			count := 0
			for message := range ch {
				if message.Err == nil && testCase.HasError {
					t.Errorf("expected failure got %v", message)
				}
				if message.Err != nil && !testCase.HasError {
					t.Errorf("expected success got: %v", message.Err)
				}
				if message.Err == nil {
					count++
				}
			}
			if count != testCase.ExpectedSentResult+testCase.ExpectedReceivedResult {
				t.Errorf("expected %v result, got %v", testCase.ExpectedSentResult+testCase.ExpectedReceivedResult, count)
			}
		})
	}
}

func TestRepo_GetReceivedMessages(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	// create test cases
	tests, sendRequests, receiveRequests := td.CreateTTsGetReceivedMessages(&conf.Database.Messages)
	// insert messages to test getting them
	messageIDs, err := insertMessages(sendRequests, receiveRequests, &conf.Database.Messages, logger)
	if err != nil {
		t.Fatalf("unable to insert message for getReceivedMessages test: %v", err)
	}
	// clean the db after the test
	defer func() {
		deleteMessagesReq := &svc.DeleteMessagesReq{
			MailboxID:  sendRequests[0].MailboxID,
			MessageIDs: messageIDs,
		}
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the getReceivedMessages request
			ch := make(chan mdl.GetMessageResponse, 1)
			go repo.GetReceivedMessages(ctx, testCase.MailboxID, ch)
			// get back the result
			count := 0
			for message := range ch {
				if message.Err == nil && testCase.HasError {
					t.Errorf("expected failure got %v", message)
				}
				if message.Err != nil && !testCase.HasError {
					t.Errorf("expected success got: %v", message.Err)
				}
				if message.Err == nil {
					count++
				}
			}
			if count != testCase.ExpectedReceivedResult {
				t.Errorf("expected %v result, got %v", testCase.ExpectedReceivedResult, count)
			}
		})
	}
}

func TestRepo_GetSentMessages(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	// create test cases
	tests, sendRequests, receiveRequests := td.CreateTTsGetSentMessages(&conf.Database.Messages)
	// insert messages to test getting them
	messageIDs, err := insertMessages(sendRequests, receiveRequests, &conf.Database.Messages, logger)
	if err != nil {
		t.Fatalf("unable to insert message for getSentMessages test: %v", err)
	}
	// clean the db after the test
	defer func() {
		deleteMessagesReq := &svc.DeleteMessagesReq{
			MailboxID:  sendRequests[0].MailboxID,
			MessageIDs: messageIDs,
		}
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the getSentMessages request
			ch := make(chan mdl.GetMessageResponse, 1)
			go repo.GetSentMessages(ctx, testCase.MailboxID, ch)
			// get back the result
			count := 0
			for message := range ch {
				if message.Err == nil && testCase.HasError {
					t.Errorf("expected failure got %v", message)
				}
				if message.Err != nil && !testCase.HasError {
					t.Errorf("expected success got: %v", message.Err)
				}
				if message.Err == nil {
					count++
				}
			}
			if count != testCase.ExpectedSentResult {
				t.Errorf("expected %v result, got %v", testCase.ExpectedSentResult, count)
			}
		})
	}
}

func TestRepo_GetMessageByParent(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	// create test cases
	tests, sendRequests, receiveRequests := td.CreateTTsGetMessagesByParent(&conf.Database.Messages)
	// insert messages to test getting them
	messageIDs, err := insertMessages(sendRequests, receiveRequests, &conf.Database.Messages, logger)
	if err != nil {
		t.Fatalf("unable to insert message for getMessagesByParent test: %v", err)
	}
	// clean the db after the test
	defer func() {
		deleteMessagesReq := &svc.DeleteMessagesReq{
			MailboxID:  sendRequests[0].MailboxID,
			MessageIDs: messageIDs,
		}
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the getMessagesByParent request
			ch := make(chan mdl.GetMessageResponse, 1)
			go repo.GetMessagesByParent(ctx, testCase.Req, ch)
			// get back the result
			count := 0
			for message := range ch {
				if message.Err == nil && testCase.HasError {
					t.Errorf("expected failure got %v", message)
				}
				if message.Err != nil && !testCase.HasError {
					t.Errorf("expected success got: %v", message.Err)
				}
				if message.Err == nil {
					count++
				}
			}
			if count != testCase.ExpectedNumber {
				t.Errorf("expected %v result, got %v", testCase.ExpectedNumber, count)
			}
		})
	}
}

func TestRepo_DeleteMessage(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	// create test cases
	tests, parentToAdd, childrenToAdd := td.CreateTTsDeleteMessage(&conf.Database.Messages)
	// insert parent message
	addRes := sendMessage(parentToAdd, &conf.Database.Messages, logger)
	if addRes.Err != nil {
		t.Fatalf("unable to insert parent message for deleteMessage test: %v", addRes.Err)
	}
	// insert 3 children messages
	ids, err := insertMessages(childrenToAdd(addRes.MessageID, 3), nil, &conf.Database.Messages, logger)
	if err != nil {
		t.Fatalf("unable to insert children message for deleteMessage test: %v", err)
	}
	// clean the db after the test in case if deletion failure
	defer func() {
		ids = append(ids, &common.MessageID{Value: addRes.MessageID})
		deleteMessagesReq := &svc.DeleteMessagesReq{
			MailboxID:  parentToAdd.MailboxID,
			MessageIDs: ids,
		}
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the deleteMessage request
			ch := make(chan error, 1)
			repo.DeleteMessage(ctx, testCase.GetReq(addRes.MessageID), ch)
			// get back the result
			err := <-ch
			// check result
			if err == nil && testCase.HasError {
				t.Errorf("expected failure got nothing")
			}
			if err != nil && !testCase.HasError {
				t.Errorf("expected success got: %v", err)
			}
		})
	}
}

func TestRepo_DeleteMessages(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	tests, addRequests := td.CreateTTsDeleteMessages(&conf.Database.Messages)
	// insert messages to test deleting them
	messageIDs, err := insertMessages(addRequests, nil, &conf.Database.Messages, logger)
	if err != nil {
		t.Fatalf("unable to insert message for deleteMessages test: %v", err)
	}
	// clean the db after the test in case of deletion failure
	defer func() {
		deleteMessagesReq := &svc.DeleteMessagesReq{
			MailboxID:  addRequests[0].MailboxID,
			MessageIDs: messageIDs,
		}
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the deleteMessages request
			ch := make(chan error, 1)
			repo.DeleteMessages(ctx, testCase.GetReq(messageIDs), ch)
			// get back the result
			err := <-ch
			if err == nil && testCase.HasError {
				t.Errorf("expected failure got %v", err)
			}
			if err != nil && !testCase.HasError {
				t.Errorf("expected success got: %v", err)
			}
		})
	}
}

func TestRepo_MarkAsRead(t *testing.T) {
	t.Parallel()
	// prepare logger and conf for repo
	logger := hlp.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unable to load configuration %v", err)
	}
	// create test cases
	tests, addRequest := td.CreateTTsMarkAsRead(&conf.Database.Messages)
	// insert a message to test marking it as read
	addRes := sendMessage(addRequest, &conf.Database.Messages, logger)
	if addRes.Err != nil {
		t.Fatalf("unable to insert message for MarkAs Read test: %v", addRes.Err)
	}
	// clean the db after the test
	defer func() {
		deleteMessagesReq := getDeleteMessagesRequest(addRequest.MailboxID, []string{addRes.MessageID})
		ensureClean(deleteMessagesReq, &conf.Database.Messages, logger)
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create the repo
			repo := dynamo.NewMessagesRepo(testCase.Conf, logger)
			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			// create a channel and submit the MarkAsRead request
			ch := make(chan error, 1)
			repo.MarkAsRead(ctx, testCase.GetReq(addRes.MessageID), ch)
			// get back the result
			err := <-ch
			// check result
			if err == nil && testCase.HasError {
				t.Errorf("expected failure got nothing")
			}
			if err != nil && !testCase.HasError {
				t.Errorf("expected success got: %v", err)
			}
		})
	}
}

// sendMessage receives SendMessageReq request and inserts a send message in db
func sendMessage(req *svc.SendMessageReq, conf *config.Messages, logger *logrus.Logger) mdl.SendMessageResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ch := make(chan mdl.SendMessageResponse, 1)
	repo := dynamo.NewMessagesRepo(conf, logger)
	repo.SendMessage(ctx, req, ch)
	res := <-ch
	return res
}

// receiveMessage receives SendMessageReq request and inserts a receive message in db
func receiveMessage(req *svc.SendMessageReq, conf *config.Messages, logger *logrus.Logger) mdl.SendMessageResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ch := make(chan mdl.SendMessageResponse, 1)
	repo := dynamo.NewMessagesRepo(conf, logger)
	repo.ReceiveMessage(ctx, req, ch)
	res := <-ch
	return res
}

// insertMessages receives o list of SendMessageReq requests and inserts the messages in db
func insertMessages(sendRequests []*svc.SendMessageReq, receiveRequests []*svc.SendMessageReq, conf *config.Messages, logger *logrus.Logger) ([]*common.MessageID, error) {
	var messageIDs []*common.MessageID
	// add send messages
	for _, req := range sendRequests {
		addRes := sendMessage(req, conf, logger)
		if addRes.Err != nil {
			return nil, addRes.Err
		}
		messageIDs = append(messageIDs, &common.MessageID{Value: addRes.MessageID})
	}
	// add receive messages
	for _, req := range receiveRequests {
		addRes := receiveMessage(req, conf, logger)
		if addRes.Err != nil {
			return nil, addRes.Err
		}
		messageIDs = append(messageIDs, &common.MessageID{Value: addRes.MessageID})
	}
	return messageIDs, nil
}

// cleanMessages receives a list of getMessage requests and delete the correspondent messages from db
func cleanMessages(req *svc.DeleteMessagesReq, conf *config.Messages, logger *logrus.Logger) error {
	if len(req.MessageIDs) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	repo := dynamo.NewMessagesRepo(conf, logger)
	ch := make(chan error, 1)
	repo.DeleteMessages(ctx, req, ch)
	if err := <-ch; err != nil {
		return err
	}
	return nil
}

// ensureClean ensures the cleanMessages function even if there is a panic
func ensureClean(req *svc.DeleteMessagesReq, conf *config.Messages, logger *logrus.Logger) {
	r := recover()
	err := cleanMessages(req, conf, logger)
	if err != nil {
		logger.Errorf("cleanUp Error: %v", err)
	}
	if r != nil {
		panic(r)
	}
}

// getDeleteMessagesRequest returns a deleteMessages request after receiving the mailboxID and the messages IDs
func getDeleteMessagesRequest(mailboxID *common.MailBoxID, messageIDs []string) *svc.DeleteMessagesReq {
	req := &svc.DeleteMessagesReq{MailboxID: mailboxID}
	for _, msgID := range messageIDs {
		req.MessageIDs = append(req.MessageIDs, &common.MessageID{Value: msgID})
	}
	return req
}
