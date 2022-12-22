package models

import (
	"time"

	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/google/uuid"
)

const (
	CORRELATION_ID    string = "CorrelationID"
	FUNCTION          string = "Function"
	MAILBOX_ID        string = "MailboxID"
	MESSAGE_ID        string = "MessageID"
	MESSAGES_NUMBER   string = "MessagesNumber"
	SENDER_ID         string = "SenderID"
	RECEIVER_ID       string = "ReceiverID"
	IS_OUT_GOING      string = "IsOutgoing"
	PARENT_MESSAGE_ID string = "ParentMessageID"
	IS_READ           string = "IsRead"
	WITH_DESCENDANTS  string = "WithDescendants"
)

type SendMessageResponse struct {
	MessageID string
	Err       error
}

type GetMessageResponse struct {
	Message MessageModel
	Err     error
}

type MailMessageInfo struct {
	SenderID        *common.UserID
	RecipientID     *common.UserID
	SenderFirstName string
	SenderLastName  string
	Title           string
	MessageBody     string
	MessageType     common.MessageType
	NoReply         bool
}

// MessageModel represents the struct of the Message in the database
type MessageModel struct {
	MailboxID       string             `json:"MailboxID"`
	MessageID       string             `json:"MessageID"`
	ParentMessageID string             `json:"ParentMessageID,omitempty"`
	SenderID        string             `json:"SenderID"`
	RecipientID     string             `json:"RecipientID"`
	SenderFirstName string             `json:"SenderFirstName"`
	SenderLastName  string             `json:"SenderLastName"`
	Title           string             `json:"Title"`
	MessageBody     string             `json:"MessageBody"`
	MessageType     common.MessageType `json:"MessageType"`
	NoReply         bool               `json:"NoReply"`
	IsRead          bool               `json:"IsRead,omitempty"`
	IsOutgoing      bool               `json:"IsOutgoing"`
	CreatedAt       uint64             `json:"CreatedAt"`
}

// NewMessageFromRequest creates and returns a MessageModel from a Send Message Request
func NewMessageFromRequest(req *svc.SendMessageReq, outgoing bool) MessageModel {
	message := MessageModel{
		MailboxID:       req.MailboxID.Value,
		MessageID:       uuid.New().String(),
		SenderID:        req.MessageInfo.SenderID.Value,
		RecipientID:     req.MessageInfo.RecipientID.Value,
		SenderFirstName: req.MessageInfo.SenderFirstName,
		SenderLastName:  req.MessageInfo.SenderLastName,
		Title:           req.MessageInfo.Title,
		MessageBody:     req.MessageInfo.MessageBody,
		MessageType:     req.MessageInfo.MessageType,
		NoReply:         req.MessageInfo.NoReply,
		IsRead:          false,
		IsOutgoing:      outgoing,
		CreatedAt:       uint64(time.Now().Unix()),
	}
	// check if parentID exist
	if req.ParentMessageID != nil {
		message.ParentMessageID = req.ParentMessageID.Value
	}
	return message
}
