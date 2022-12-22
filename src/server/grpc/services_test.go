package grpc_test

import (
	"context"
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/database/dynamo"
	mdl "github.com/fatmalabidi/MailBoxServcie/src/models"
	"github.com/fatmalabidi/MailBoxServcie/src/server/grpc/grpcTestUtils"
	"github.com/fatmalabidi/MailBoxServcie/src/testdata/service"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"testing"
	"time"
)

const LIMIT_TESTDATA = 20

func TestSvcHandler_SendMessage(t *testing.T) {
	t.Parallel()
	var logger = helpers.GetTestLogger()
	var svr = grpcTestUtils.StartGRPCServer(logger)
	var getReqs []*svc.GetMessageReq

	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()

	// Tests for the cases of existing parentID
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config")
	}

	defer func() {
		cleanup(conf, logger, getReqs)
	}()

	mailboxID := &common.MailBoxID{Value: uuid.New().String()}
	parentRequestNoReplay, parentRequestReplay := service.CreateParentReq(mailboxID)

	parentIDs, err := sendMessages(conf, logger, []*svc.SendMessageReq{parentRequestNoReplay, parentRequestReplay})
	parentIDNoReplay := parentIDs[0]
	parentIDWithReplay := parentIDs[1]
	getReqs = append(getReqs, &svc.GetMessageReq{
		MailboxID: mailboxID,
		MessageID: &common.MessageID{Value: parentIDNoReplay},
	}, &svc.GetMessageReq{
		MailboxID: mailboxID,
		MessageID: &common.MessageID{Value: parentIDWithReplay},
	})

	testCases := service.CreateTTSendMessage(parentIDNoReplay, parentIDWithReplay, mailboxID)
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx, _ := helpers.AddTimeoutToCtx(context.Background(), 10)
			res, err := client.SendMessage(ctx, testCase.Req)
			if !testCase.HasError && err == nil {
				// in the happy-path the message has been successfully added so we need to remove it in the cleanup func
				getReqs = append(getReqs, &svc.GetMessageReq{
					MailboxID: &common.MailBoxID{Value: testCase.Req.MailboxID.Value},
					MessageID: &common.MessageID{Value: res.Value},
				})
			}
			if testCase.HasError && err == nil {
				t.Error("expected to get error, got nil")
				getReqs = append(getReqs, &svc.GetMessageReq{
					MailboxID: &common.MailBoxID{Value: testCase.Req.MailboxID.Value},
					MessageID: &common.MessageID{Value: res.Value},
				})
			}
			if !testCase.HasError && err != nil {
				t.Errorf("expected success, got error %v", err)
			}
		})
	}

	testCase := testCases[0]
	t.Run("invalid context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		res, err := client.SendMessage(ctx, testCase.Req)
		if err == nil {
			getReqs = append(getReqs, &svc.GetMessageReq{
				MailboxID: &common.MailBoxID{Value: testCase.Req.MailboxID.Value},
				MessageID: &common.MessageID{Value: res.Value},
			})
			t.Error("expected to get error invalid context, got nil")
		}
	})
}

func TestSvcHandler_GetMessage(t *testing.T) {
	t.Parallel()
	// prepare logger, config and run server
	logger := helpers.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config for GetMessage test")
	}
	svr := grpcTestUtils.StartGRPCServer(logger)
	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()
	// create test cases
	tests, messageToAdd := service.CreateTTGetMessage()
	// insert a message to test getting it
	ids, err := sendMessages(conf, logger, []*svc.SendMessageReq{messageToAdd})
	if err != nil || len(ids) == 0 {
		t.Fatal("unable to insert message for GetMessage test")
	}
	// clean up the db after the test
	defer func() {
		cleanup(conf, logger, []*svc.GetMessageReq{{
			MailboxID: messageToAdd.MailboxID,
			MessageID: &common.MessageID{Value: ids[0]},
		}})
	}()
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx, cancel := helpers.AddTimeoutToCtx(context.Background(), testCase.ContextDeadline)
			defer cancel()
			_, err := client.GetMessage(ctx, testCase.GetReq(ids[0]))
			if err == nil && testCase.HasError {
				t.Error("expected to get error, got nil")
			}
			if err != nil && !testCase.HasError {
				t.Errorf("expected success, got error %v", err)
			}
		})
	}
}

func TestSvcHandler_GetMessages(t *testing.T) {
	t.Parallel()
	// prepare logger, config and run server
	logger := helpers.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config for GetMessage test")
	}
	svr := grpcTestUtils.StartGRPCServer(logger)
	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()

	var mailboxID = &common.MailBoxID{Value: uuid.New().String()}
	var messagesToSend = service.CreateSendMessageReqs(mailboxID, LIMIT_TESTDATA)
	// insert a message to test getting it
	ids, err := sendMessages(conf, logger, messagesToSend)
	if err != nil || len(ids) == 0 {
		t.Fatal("unable to insert message for GetMessages test")
	}
	// clean up the db after the test
	defer func() {
		getReqs := service.CreateGetMessageReqs(mailboxID, ids)
		cleanup(conf, logger, getReqs)
	}()

	testCases := service.CreateTTGetMessages(mailboxID, LIMIT_TESTDATA)
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx, cancel := helpers.AddTimeoutToCtx(context.Background(), 10)
			defer cancel()
			stream, err := client.GetMessages(ctx, testCase.Id)
			if err != nil {
				t.Fatal("expected success got error", err)
			}
			messageCount := 0
			for {
				res, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if testCase.HasError && err == nil {
					t.Error("expected error got nil")
				}
				if !testCase.HasError && err != nil && err != io.EOF {
					t.Error("expected success got error ", err)
				}
				if res != nil && !testCase.HasError {
					messageCount++
				}
				if err != nil && err != io.EOF {
					break
				}
			}
			if messageCount != testCase.ExpectedMessages {
				t.Errorf("expected messageCount= %d got %d ", testCase.ExpectedMessages, messageCount)
			}
		})

		testCase := testCases[0]
		t.Run("invalid context", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*conf.Server.Deadline*time.Second)
			defer cancel()
			stream, err := client.GetMessages(ctx, testCase.Id)
			if err != nil {
				t.Fatal("unable create the stream")
				return
			}
			_, err = stream.Recv()
			if err == nil {
				t.Error("expected to get error invalid context, got nil")
			}
		})
	}
}

func TestSvcHandler_MarkAsRead(t *testing.T) {
	t.Parallel()
	// prepare logger, config and run server
	logger := helpers.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config for MarkAsRead test")
	}
	svr := grpcTestUtils.StartGRPCServer(logger)
	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()

	var mailboxID = &common.MailBoxID{Value: uuid.New().String()}
	var messagesToSend = service.CreateSendMessageReqs(mailboxID, 1)
	// insert a message to test getting it
	ids, err := sendMessages(conf, logger, messagesToSend)
	if err != nil || len(ids) == 0 {
		t.Fatal("unable to insert message for MarkAsRead test")
	}
	messageID := &common.MessageID{Value: ids[0]}
	// clean up the db after the test
	defer func() {
		getReqs := service.CreateGetMessageReqs(mailboxID, ids)
		cleanup(conf, logger, getReqs)
	}()
	testCases := service.CreateTTMarkAsRead(mailboxID, messageID)
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), conf.Server.Deadline*time.Second)
			defer cancel()
			_, err := client.MarkAsRead(ctx, testCase.Req)
			if err == nil && testCase.HasError {
				t.Error("expected to get error, got nil")
			}
			if err != nil && !testCase.HasError {
				t.Errorf("expected success, got error %v", err)
			}
		})
	}

	// 1st testCase is the happy path
	testCase := testCases[0]
	t.Run("invalid context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*conf.Server.Deadline*time.Second)
		defer cancel()
		_, err := client.MarkAsRead(ctx, testCase.Req)
		if err == nil {
			t.Error("expected to get error invalid context, got nil")
		}
	})

}

func TestSvcHandler_GetSentMessages(t *testing.T) {
	t.Parallel()
	// prepare logger, config and run server
	logger := helpers.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config for GetSentMessages test")
	}
	svr := grpcTestUtils.StartGRPCServer(logger)
	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()
	// create test cases
	tests, sentMessageToAdd, receivedMessagesToAdd := service.CreateTTGetSentMessages()
	// clean up the db after the test
	var sentIds, receivedIds []string
	defer func() {
		sentMessagesToDelete := service.CreateGetMessageReqs(sentMessageToAdd[0].MailboxID, sentIds)
		receivedMessagesToDelete := service.CreateGetMessageReqs(receivedMessagesToAdd[0].MailboxID, receivedIds)
		cleanup(conf, logger, append(sentMessagesToDelete, receivedMessagesToDelete...))
	}()
	// insert the messages to test getting them
	sentIds, err = sendMessages(conf, logger, sentMessageToAdd)
	if err != nil || len(sentIds) == 0 {
		t.Fatal("unable to insert sent message for GetSentMessages test")
	}
	receivedIds, err = receiveMessages(conf, logger, receivedMessagesToAdd)
	if err != nil || len(receivedIds) == 0 {
		t.Fatal("unable to insert received message for GetSentMessages test")
	}
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create a context with timeout
			ctx, cancel := helpers.AddTimeoutToCtx(context.Background(), testCase.ContextDeadline)
			defer cancel()
			// submit request
			stream, err := client.GetSentMessages(ctx, testCase.MailboxID)
			if err != nil {
				t.Fatalf("expected success got error, %v", err)
			}
			// loadedMessages counts the number of returned messages
			receivedMessages := 0
			for {
				res, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err == nil && testCase.HasError {
					t.Errorf("expected error got %v", res)
				}
				if err != nil && err != io.EOF && !testCase.HasError {
					t.Error("expected success got error ", err)
				}
				if res != nil {
					receivedMessages++
				}
				if err != nil {
					break
				}
			}
			if receivedMessages != testCase.ExpectedResults {
				t.Errorf("expected %d message, got %d", testCase.ExpectedResults, receivedMessages)
			}
		})
	}
}

func TestSvcHandler_GetMessageByParent(t *testing.T) {
	t.Parallel()
	// prepare logger, config and run server
	logger := helpers.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config for GetMessagesByParent test")
	}
	svr := grpcTestUtils.StartGRPCServer(logger)
	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()
	// create test cases
	tests, sentMessageToAdd, receivedMessagesToAdd := service.CreateTTGetMessagesByParent()
	// clean up the db after the test
	var sentIds, receivedIds []string
	defer func() {
		sentMessagesToDelete := service.CreateGetMessageReqs(sentMessageToAdd[0].MailboxID, sentIds)
		receivedMessagesToDelete := service.CreateGetMessageReqs(receivedMessagesToAdd[0].MailboxID, receivedIds)
		cleanup(conf, logger, append(sentMessagesToDelete, receivedMessagesToDelete...))
	}()
	// insert the messages to test getting them
	sentIds, err = sendMessages(conf, logger, sentMessageToAdd)
	if err != nil || len(sentIds) == 0 {
		t.Fatal("unable to insert sent message for GetMessagesByParent test")
	}
	receivedIds, err = receiveMessages(conf, logger, receivedMessagesToAdd)
	if err != nil || len(receivedIds) == 0 {
		t.Fatal("unable to insert received message for GetMessagesByParent test")
	}
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			// create a context with timeout
			ctx, cancel := helpers.AddTimeoutToCtx(context.Background(), testCase.ContextDeadline)
			defer cancel()
			// submit request
			stream, err := client.GetMessagesByParent(ctx, testCase.Req)
			if err != nil {
				t.Fatalf("expected success got error, %v", err)
			}
			// loadedMessages counts the number of returned messages
			receivedMessages := 0
			for {
				res, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err == nil && testCase.HasError {
					t.Errorf("expected error got %v", res)
				}
				if err != nil && err != io.EOF && !testCase.HasError {
					t.Error("expected success got error ", err)
				}
				if res != nil {
					receivedMessages++
				}
				if err != nil {
					break
				}
			}
			if receivedMessages != testCase.ExpectedMessages {
				t.Errorf("expected %d message, got %d", testCase.ExpectedMessages, receivedMessages)
			}
		})
	}
}

func TestSvcHandler_GetReceivedMessages(t *testing.T) {

	t.Parallel()
	// prepare logger, config and run server
	logger := helpers.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config for GetReceivedMessages test")
	}
	svr := grpcTestUtils.StartGRPCServer(logger)
	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()
	var mailboxID = &common.MailBoxID{Value: uuid.New().String()}
	receivedMessagesReq := service.CreateSendMessageReqs(mailboxID, LIMIT_TESTDATA)
	ids, err := receiveMessages(conf, logger, receivedMessagesReq)
	if err != nil {
		t.Fatal("unable create receivedMessages")
	}
	defer func() {
		getReqs := service.CreateGetMessageReqs(mailboxID, ids)
		cleanup(conf, logger, getReqs)
	}()

	testCases := service.CreateTTGetReceivedMessages(mailboxID, LIMIT_TESTDATA)
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), conf.Server.Deadline*time.Second)
			defer cancel()
			stream, err := client.GetReceivedMessages(ctx, testCase.Id)
			if err != nil {
				t.Fatal("expected success got error", err)
			}
			// the count of received messages
			messageCount := 0
			for {
				res, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if testCase.HasError && err == nil {
					t.Error("expected error got nil")
				}
				if !testCase.HasError && err != nil && err != io.EOF {
					t.Error("expected success got error ", err)
				}
				if res != nil && !testCase.HasError {
					messageCount++
				}
				if err != nil {
					break
				}
			}
			if messageCount != testCase.ExpectedMessages {
				t.Errorf("expected messageCount= %d got %d ", testCase.ExpectedMessages, messageCount)
			}
		})
	}

	// 1st testcase is the happy path
	testCase := testCases[0]
	t.Run("invalid context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*conf.Server.Deadline*time.Second)
		defer cancel()
		stream, err := client.GetReceivedMessages(ctx, testCase.Id)
		if err != nil {
			t.Fatal("unable create the stream")
			return
		}
		_, err = stream.Recv()
		if err == nil {
			t.Error("expected to get error invalid context, got nil")
		}
	})
}

func TestSvcHandler_DeleteMessage(t *testing.T) {
	t.Parallel()
	// prepare logger, config and run server
	logger := helpers.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config for DeleteMessage test")
	}
	svr := grpcTestUtils.StartGRPCServer(logger)
	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()
	// clean up the db after the test
	var parentIDs, childrenIDs []string
	var getReqs []*svc.GetMessageReq
	defer func() {
		cleanup(conf, logger, getReqs)
	}()
	// create test cases
	tests, parentToAdd, childrenToAdd := service.CreateTTDeleteMessage()
	// insert a parent message to test deleting it
	parentIDs, err = sendMessages(conf, logger, []*svc.SendMessageReq{parentToAdd})
	if err != nil || len(parentIDs) == 0 {
		t.Fatal("unable to insert message for DeleteMessage test")
	}
	// insert 3 children messages to test deleting them with the parent
	childrenIDs, err = sendMessages(conf, logger, childrenToAdd(parentIDs[0], 3))
	if err != nil || len(childrenIDs) == 0 {
		t.Fatal("unable to insert message for DeleteMessage test")
	}
	getReqs = service.CreateGetMessageReqs(parentToAdd.MailboxID, append(parentIDs, childrenIDs...))
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx, cancel := helpers.AddTimeoutToCtx(context.Background(), testCase.ContextDeadline)
			defer cancel()
			_, err := client.DeleteMessage(ctx, testCase.GetReq(parentIDs[0]))
			if err == nil && testCase.HasError {
				t.Error("expected to get error, got nil")
			}
			if err != nil && !testCase.HasError {
				t.Errorf("expected success, got error %v", err)
			}
		})
	}
}

func TestSvcHandler_DeleteMessages(t *testing.T) {
	t.Parallel()
	// prepare logger, config and run server
	logger := helpers.GetTestLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fatal("unable to load config for DeleteMessages test")
	}
	svr := grpcTestUtils.StartGRPCServer(logger)
	// create client
	client, conn := svr.StartClientConnection()
	// close the connection
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("expected to successfully close the connection, got %v ", err)
		}
	}()
	// clean up the db after the test
	var getReqs []*svc.GetMessageReq
	defer func() {
		cleanup(conf, logger, getReqs)
	}()
	// create test cases
	tests, messagesToAdd := service.CreateTTDeleteMessages()
	// insert list of messages to test deleting them
	insertedIDs, err := sendMessages(conf, logger, messagesToAdd)
	if err != nil || len(insertedIDs) == 0 {
		t.Fatal("unable to insert message for DeleteMessages test")
	}
	getReqs = service.CreateGetMessageReqs(messagesToAdd[0].MailboxID, insertedIDs)
	// run through all test cases
	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx, cancel := helpers.AddTimeoutToCtx(context.Background(), testCase.ContextDeadline)
			defer cancel()
			_, err := client.DeleteMessages(ctx, testCase.GetReq(insertedIDs))
			if err == nil && testCase.HasError {
				t.Error("expected to get error, got nil")
			}
			if err != nil && !testCase.HasError {
				t.Errorf("expected success, got error %v", err)
			}
		})
	}
}

// cleanup removes messages that have been added in the test
func cleanup(conf *config.Config, logger *logrus.Logger, reqs []*svc.GetMessageReq) {
	r := recover()
	for _, req := range reqs {
		ch := make(chan error, 1)
		repo := dynamo.NewMessagesRepo(&conf.Database.Messages, logger)
		delReq := &svc.DeleteMessageReq{
			MailboxID: req.MailboxID,
			MessageID: req.MessageID,
		}
		repo.DeleteMessage(context.Background(), delReq, ch)
		err := <-ch
		if err != nil {
			logger.Error("unable cleanup data")
		}
		if r != nil {
			panic(r)
		}
	}
}

// sendMessages sends messages and returns the list of the sent messages IDs
func sendMessages(conf *config.Config, logger *logrus.Logger, reqs []*svc.SendMessageReq) ([]string, error) {
	repo := dynamo.NewMessagesRepo(&conf.Database.Messages, logger)
	var ids []string
	for _, req := range reqs {
		sendMsgCh := make(chan mdl.SendMessageResponse, 1)
		repo.SendMessage(context.Background(), req, sendMsgCh)
		res := <-sendMsgCh
		if res.Err != nil {
			return nil, res.Err
		}
		ids = append(ids, res.MessageID)
	}
	return ids, nil
}

// receiveMessages receives messages and returns the list of the received messages IDs
func receiveMessages(conf *config.Config, logger *logrus.Logger, reqs []*svc.SendMessageReq) ([]string, error) {
	repo := dynamo.NewMessagesRepo(&conf.Database.Messages, logger)
	var ids []string
	for _, req := range reqs {
		sendMsgCh := make(chan mdl.SendMessageResponse, 1)
		repo.ReceiveMessage(context.Background(), req, sendMsgCh)
		res := <-sendMsgCh
		if res.Err != nil {
			return nil, res.Err
		}
		ids = append(ids, res.MessageID)
	}
	return ids, nil
}
