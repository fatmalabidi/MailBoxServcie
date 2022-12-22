package grpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/database"
	mdl "github.com/fatmalabidi/MailBoxServcie/src/models"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/converters"
	hlp "github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	vld "github.com/fatmalabidi/MailBoxServcie/src/utils/validators"
	"github.com/fatmalabidi/protobuf/common"
	"github.com/fatmalabidi/protobuf/mailbox"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SvcHandler contains all necessary components handlers and configuration related to the server
type SvcHandler struct {
	DB     database.MessagesHandler
	Config *config.Config
	Logger *logrus.Logger
}

func NewSvcHandler(conf *config.Config, logger *logrus.Logger) *SvcHandler {
	db, err := database.CreateMessagesHandler(&conf.Database.Messages, logger)
	if err != nil {
		logger.Panicf("Fail to create db handler :%v", err)
	}
	return &SvcHandler{
		DB:     db,
		Config: conf,
		Logger: logger,
	}

}

func (srv *SvcHandler) SendMessage(ctx context.Context, req *svc.SendMessageReq) (*common.MessageID, error) {
	// TODO add USER_ID after adding auth
	logger := hlp.EnrichLog(srv.Logger, ctx, map[string]string{
		mdl.FUNCTION:    "SendMessage",
		mdl.MAILBOX_ID:  req.MailboxID.Value,
		mdl.SENDER_ID:   req.MessageInfo.SenderID.Value,
		mdl.RECEIVER_ID: req.MessageInfo.RecipientID.Value,
	})

	// check if this message has a message parent and its parent has np-reply=true/false
	if req.ParentMessageID != nil {
		if !vld.IsValidID(req.ParentMessageID.Value) {
			logger.Error("invalid parentID format")
			return nil, status.Error(codes.InvalidArgument, "invalid parentID format")
		}
		getCh := make(chan mdl.GetMessageResponse, 1)
		getReq := &svc.GetMessageReq{
			MailboxID: req.MailboxID,
			MessageID: req.ParentMessageID,
		}
		srv.DB.GetMessage(ctx, getReq, getCh)
		var getRes mdl.GetMessageResponse
		select {
		// check context exceeded error
		case <-ctx.Done():
			return nil, status.Error(codes.DeadlineExceeded, ctx.Err().Error())
		case getRes = <-getCh:
			break
		}
		if getRes.Err != nil {
			logger.WithError(getRes.Err).Error("error getting message parent")
			return nil, status.Errorf(codes.Internal, "error getting message parent,%v", getRes.Err.Error())
		}
		if getRes.Message.NoReply {
			logger.Error("the parent message doesn't allow reply")
			return nil, status.Error(codes.FailedPrecondition, "the parent message doesn't allow reply")
		}
	}
	sendCh := make(chan mdl.SendMessageResponse, 1)
	go srv.DB.SendMessage(ctx, req, sendCh)
	var sendRes mdl.SendMessageResponse
	select {
	// check context exceeded error
	case <-ctx.Done():
		logger.WithError(ctx.Err()).Error(ctx.Err().Error())
		return nil, status.Error(codes.DeadlineExceeded, ctx.Err().Error())
	case sendRes = <-sendCh:
		break
	}
	if sendRes.Err != nil {
		logger.WithError(sendRes.Err).Error("error when sending the message")
		return nil, status.Errorf(codes.Internal, "error when sending the message %v", sendRes.Err.Error())
	}
	logger.WithField(mdl.MESSAGE_ID, sendRes.MessageID).Info("message sent successfully")
	return &common.MessageID{Value: sendRes.MessageID}, nil
}

func (srv *SvcHandler) GetMessage(ctx context.Context, req *svc.GetMessageReq) (*mailbox.MailMessage, error) {
	// set logger
	logger := hlp.EnrichLog(srv.Logger, ctx, map[string]string{
		mdl.FUNCTION:   "GetMessage",
		mdl.MAILBOX_ID: req.MailboxID.Value,
		mdl.MESSAGE_ID: req.MessageID.Value,
	})
	// create channel and submit the request
	getChan := make(chan mdl.GetMessageResponse, 1)
	go srv.DB.GetMessage(ctx, req, getChan)
	// get back the results
	var getResult mdl.GetMessageResponse
	select {
	// receive context error
	case <-ctx.Done():
		logger.WithError(ctx.Err()).Error("failure due context")
		return nil, status.Error(codes.DeadlineExceeded, ctx.Err().Error())
	// receive result getResult from channel
	case getResult = <-getChan:
		break
	}
	// check error
	if getResult.Err != nil {
		logger.WithError(getResult.Err).Error("can not get the message")
		return nil, status.Error(codes.Internal, getResult.Err.Error())
	}
	// convert the received message to proto MailMessage
	mailMessage := converters.ConvertMessageModelToProto(getResult.Message)
	// return the result
	logger.Info("Message got with success")
	return mailMessage, nil
}

func (srv *SvcHandler) GetMessages(mailboxID *common.MailBoxID, server svc.MailboxService_GetMessagesServer) error {

	logger := hlp.EnrichLog(srv.Logger, server.Context(), map[string]string{
		mdl.FUNCTION:   "GetMessages",
		mdl.MAILBOX_ID: mailboxID.Value,
	}) // check context
	ch := make(chan mdl.GetMessageResponse, 1)
	go srv.DB.GetMessages(server.Context(), mailboxID, ch)
	count := 0
forLoop:
	for {
		select {
		// check context exceeded error
		case <-server.Context().Done():
			logger.WithError(server.Context().Err()).Error()
			return status.Error(codes.DeadlineExceeded, server.Context().Err().Error())
		case res, ok := <-ch:
			if !ok {
				break forLoop
			}
			count++
			if res.Err != nil {
				logger.WithError(res.Err).Error()
				return status.Error(codes.Internal, res.Err.Error())
			}
			message := converters.ConvertMessageModelToProto(res.Message)
			err := server.Send(message)
			if err != nil {
				logger.WithError(err).Error("fail to send message")
			}
		}
	}

	logger.Info(count, "messages are fetched with success")
	return nil
}

func (srv *SvcHandler) GetSentMessages(id *common.MailBoxID, serverStream svc.MailboxService_GetSentMessagesServer) error {

	// set logger
	logger := hlp.EnrichLog(srv.Logger, serverStream.Context(), map[string]string{
		mdl.FUNCTION:   "GetSentMessages",
		mdl.MAILBOX_ID: id.Value,
	})
	// create channel and submit the request
	getChan := make(chan mdl.GetMessageResponse, 1)
	go srv.DB.GetSentMessages(serverStream.Context(), id, getChan)
	// get back the results
	numberOfMessages := 0 // numberOfMessages counts the number of found messages
	// loop the channel
chanLoop:
	for {
		// check context timeout
		select {
		case <-serverStream.Context().Done():
			logger.WithError(serverStream.Context().Err()).Error()
			return status.Error(codes.DeadlineExceeded, serverStream.Context().Err().Error())
		// receive result from db
		case res, ok := <-getChan:
			if !ok {
				break chanLoop
			}
			// check error
			if res.Err != nil {
				logger.WithError(res.Err).Error()
				return status.Error(codes.Internal, res.Err.Error())
			}
			numberOfMessages++
			// convert the received message to proto MailMessage
			mailMessage := converters.ConvertMessageModelToProto(res.Message)
			// send the result
			err := serverStream.Send(mailMessage)
			if err != nil {
				logger.WithError(err).Error("fail to send message")
			}
		}
	}
	logger.Infof("%d messages have been returned with success", numberOfMessages)
	return nil
}

func (srv *SvcHandler) GetReceivedMessages(mailboxID *common.MailBoxID, server svc.MailboxService_GetReceivedMessagesServer) error {
	logger := hlp.EnrichLog(srv.Logger, server.Context(), map[string]string{
		mdl.FUNCTION:   "GetReceivedMessages",
		mdl.MAILBOX_ID: mailboxID.Value,
	})

	ch := make(chan mdl.GetMessageResponse, 1)
	go srv.DB.GetReceivedMessages(server.Context(), mailboxID, ch)
	count := 0
forLoop:
	for {
		select {
		// check context exceeded error
		case <-server.Context().Done():
			logger.WithError(server.Context().Err()).Error()
			return status.Error(codes.DeadlineExceeded, server.Context().Err().Error())
		case res, ok := <-ch:
			if !ok {
				break forLoop
			}
			if res.Err != nil {
				logger.WithError(res.Err).Error()
				return status.Error(codes.Internal, res.Err.Error())
			}
			count++
			message := converters.ConvertMessageModelToProto(res.Message)
			err := server.Send(message)
			if err != nil {
				logger.WithError(err).Error()
			}
		}
	}

	logger.Info(count, "received messages are fetched with success")
	return nil
}

func (srv *SvcHandler) DeleteMessage(ctx context.Context, req *svc.DeleteMessageReq) (*common.GenericResponse, error) {
	// set logger
	logger := hlp.EnrichLog(srv.Logger, ctx, map[string]string{
		mdl.FUNCTION:         "DeleteMessage",
		mdl.MAILBOX_ID:       req.MailboxID.Value,
		mdl.MESSAGE_ID:       req.MessageID.Value,
		mdl.WITH_DESCENDANTS: strconv.FormatBool(req.WithDescendants),
	})
	// create channel and submit the delete message request
	delChan := make(chan error, 1)
	go srv.DB.DeleteMessage(ctx, req, delChan)
	// get back the results
	var delErr error
	select {
	// chek if context is exceeded
	case <-ctx.Done():
		logger.WithError(ctx.Err()).Error()
		return nil, status.Error(codes.DeadlineExceeded, ctx.Err().Error())
	// get the result from the channel
	case delErr = <-delChan:
		break
	}
	if delErr != nil {
		logger.WithError(delErr).Error("failed to delete the message")
		return nil, status.Errorf(codes.Internal, "failed to delete the message %v", delErr.Error())
	}
	logger.Info("message successfully deleted")

	return &common.GenericResponse{
		Id:     req.MessageID.Value,
		Status: common.Status_SUCCESS,
	}, nil
}

func (srv *SvcHandler) DeleteMessages(ctx context.Context, req *svc.DeleteMessagesReq) (*common.GenericResponse, error) {
	// set logger
	logger := hlp.EnrichLog(srv.Logger, ctx, map[string]string{
		mdl.FUNCTION:        "DeleteMessages",
		mdl.MAILBOX_ID:      req.MailboxID.Value,
		mdl.MESSAGES_NUMBER: fmt.Sprintf("%d messages", len(req.MessageIDs)),
	})
	// create channel and submit the delete messages request
	delChan := make(chan error, 1)
	go srv.DB.DeleteMessages(ctx, req, delChan)
	// get back the results
	var delErr error
	select {
	// chek if context is exceeded
	case <-ctx.Done():
		logger.WithError(ctx.Err()).Error()
		return nil, status.Error(codes.DeadlineExceeded, ctx.Err().Error())
	// get the result from the channel
	case delErr = <-delChan:
		break
	}
	if delErr != nil {
		logger.WithError(delErr).Error("failed to delete the messages")
		return nil, status.Errorf(codes.Internal, "failed to delete the messages %v", delErr.Error())
	}
	logger.Info("messages successfully deleted")

	return &common.GenericResponse{
		Id:     req.MailboxID.Value,
		Status: common.Status_SUCCESS,
	}, nil
}

func (srv *SvcHandler) GetMessagesByParent(req *svc.GetMessagesByParentReq, serverStream svc.MailboxService_GetMessagesByParentServer) error {
	// set logger
	logger := hlp.EnrichLog(srv.Logger, serverStream.Context(), map[string]string{
		mdl.FUNCTION:          "GetMessageByParent",
		mdl.MAILBOX_ID:        req.MailboxID.Value,
		mdl.PARENT_MESSAGE_ID: req.ParentMessageID.Value,
	})
	// create channel and submit the request
	getChan := make(chan mdl.GetMessageResponse, 1)
	go srv.DB.GetMessagesByParent(serverStream.Context(), req, getChan)
	// get back the results
	numberOfMessages := 0 // numberOfMessages counts the number of found messages
	// loop the channel
chanLoop:
	for {
		// check context timeout
		select {
		case <-serverStream.Context().Done():
			logger.WithError(serverStream.Context().Err()).Error()
			return status.Error(codes.DeadlineExceeded, serverStream.Context().Err().Error())
		// receive result from db
		case res, ok := <-getChan:
			if !ok {
				break chanLoop
			}
			// check error
			if res.Err != nil {
				logger.WithError(res.Err).Error()
				return status.Error(codes.Internal, res.Err.Error())
			}
			numberOfMessages++
			// convert the received message to proto MailMessage
			mailMessage := converters.ConvertMessageModelToProto(res.Message)
			// send the result
			err := serverStream.Send(mailMessage)
			if err != nil {
				logger.WithError(err).Error("fail to send message")
			}
		}
	}
	logger.Infof("%d messages have been returned with success", numberOfMessages)
	return nil
}

func (srv *SvcHandler) MarkAsRead(ctx context.Context, req *svc.GetMessageReq) (*common.GenericResponse, error) {
	// TODO add USER_ID after adding auth
	logger := hlp.EnrichLog(srv.Logger, ctx, map[string]string{
		mdl.FUNCTION:   "MarkAsRead",
		mdl.MAILBOX_ID: req.MailboxID.Value,
		mdl.MESSAGE_ID: req.MessageID.Value,
	})
	ch := make(chan error, 1)
	srv.DB.MarkAsRead(ctx, req, ch)
	var resErr error
	select {
	case <-ctx.Done():
		logger.WithError(ctx.Err()).Error()
		return nil, status.Error(codes.DeadlineExceeded, ctx.Err().Error())
	case resErr = <-ch:
		break
	}
	if resErr != nil {
		logger.WithError(resErr).Error("error when marking message as read ")
		return nil, status.Errorf(codes.Internal, "error when marking message as read %v", resErr.Error())
	}
	logger.Info("messages successfully marked as read")

	return &common.GenericResponse{
		Id:     req.MessageID.Value,
		Status: common.Status_SUCCESS,
	}, nil

}
