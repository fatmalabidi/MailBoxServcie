package converters

import (
	"github.com/fatmalabidi/MailBoxServcie/src/models"
	"github.com/fatmalabidi/protobuf/common"
	"github.com/fatmalabidi/protobuf/mailbox"
)

// ConvertMessageModelToProto converts a Message model to protobuf MailMessage
func ConvertMessageModelToProto(model models.MessageModel) *mailbox.MailMessage {
	return &mailbox.MailMessage{
		MailboxID:       &common.MailBoxID{Value: model.MailboxID},
		MessageID:       &common.MessageID{Value: model.MessageID},
		ParentMessageID: &common.MessageID{Value: model.ParentMessageID},
		MailboxInfo: &mailbox.MailMessageInfo{
			SenderID:        &common.UserID{Value: model.SenderID},
			RecipientID:     &common.UserID{Value: model.RecipientID},
			SenderFirstName: model.SenderFirstName,
			SenderLastName:  model.SenderLastName,
			Title:           model.Title,
			MessageBody:     model.MessageBody,
			MessageType:     model.MessageType,
			NoReply:         model.NoReply,
		},
		IsRead:     model.IsRead,
		IsOutgoing: model.IsOutgoing,
		CreatedAt:  model.CreatedAt,
	}
}
