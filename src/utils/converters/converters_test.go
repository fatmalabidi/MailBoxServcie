package converters_test

import (
	"github.com/bxcodec/faker/v3"
	"github.com/fatmalabidi/MailBoxServcie/src/models"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/converters"
	"github.com/fatmalabidi/protobuf/mailbox"
	"testing"
)

func TestConvertMessageModelToProto(t *testing.T) {
	// generate a messageModel
	messageModel := &models.MessageModel{}
	_ = faker.FakeData(messageModel)
	// convert it to proto MailMessage
	mailMessage := converters.ConvertMessageModelToProto(*messageModel)
	if !isEqualMessageModelToProto(*messageModel, mailMessage) {
		t.Error("expected success conversion, got different object")
	}
}

// isEqualMessageModelToProto compares the fields of a MessageModel and a proto MailMessage
// it returns true if they are similar and false otherwise
func isEqualMessageModelToProto(model models.MessageModel, proto *mailbox.MailMessage) bool {
	return model.MailboxID == proto.MailboxID.Value &&
		model.MessageID == proto.MessageID.Value &&
		model.ParentMessageID == proto.ParentMessageID.Value &&
		model.SenderID == proto.MailboxInfo.SenderID.Value &&
		model.RecipientID == proto.MailboxInfo.RecipientID.Value &&
		model.SenderFirstName == proto.MailboxInfo.SenderFirstName &&
		model.Title == proto.MailboxInfo.Title &&
		model.MessageBody == proto.MailboxInfo.MessageBody &&
		model.MessageType == proto.MailboxInfo.MessageType &&
		model.IsRead == proto.IsRead &&
		model.IsOutgoing == proto.IsOutgoing &&
		model.CreatedAt == proto.CreatedAt
}
