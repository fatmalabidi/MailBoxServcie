package mailboxHandler

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	mdl "github.com/fatmalabidi/MailBoxServcie/lambda/models"
	"github.com/fatmalabidi/MailBoxServcie/src/models"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Db     *dynamodb.DynamoDB
	Logger *logrus.Logger
}

func (h *Handler) Handle(ctx context.Context, evt events.DynamoDBEvent) error {
	for _, record := range evt.Records {
		if record.EventName != "INSERT" {
			return nil
		}
		message := models.MessageModel{}
		err := unmarshal(record.Change.NewImage, &message)
		if err != nil {
			h.Logger.
				WithField("record", &record).
				WithError(err).Errorf("unable to deserialize message")
			return nil
		}
		// handle only records related to outgoing messages
		if message.IsOutgoing {
			newMessage := message
			mailboxID, err := h.getMailboxID(newMessage.RecipientID)
			if err != nil {
				switch err.(type) {
				case NotFoundError:
					return nil
				default:
					h.Logger.
						WithError(err).
						WithField("mailboxID", message.MailboxID).
						WithField("messageID", message.MessageID).
						WithField("receiverID", message.RecipientID).
						WithField("senderID", message.SenderID)
					return err
				}
			}
			newMessage.MailboxID = mailboxID
			newMessage.IsOutgoing = false
			err = h.putMessage(ctx, newMessage)
			if err != nil {
				h.Logger.
					WithError(err).
					WithField("mailboxID", message.MailboxID).
					WithField("messageID", message.MessageID).
					WithField("receiverID", message.RecipientID).
					WithField("senderID", message.SenderID)
				return err
			}
		}
	}
	return nil
}

func (h *Handler) putMessage(ctx context.Context, message models.MessageModel) error {
	dynamoMsg, _ := dynamodbattribute.MarshalMap(message)
	input := dynamodb.PutItemInput{
		Item:      dynamoMsg,
		TableName: aws.String(os.Getenv("MAILBOX_TABLE")),
	}
	_, err := h.Db.PutItemWithContext(ctx, &input)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) getMailboxID(profileID string) (string, error) {
	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				S: aws.String(profileID),
			},
		},
		ProjectionExpression: aws.String("MailBoxID"),
		TableName:            aws.String(os.Getenv("PROFILE_TABLE")),
	}
	getRes, err := h.Db.GetItem(getInput)
	if err != nil {
		return "", err
	}
	if getRes.Item == nil {
		return "", NotFoundError{Msg: "profile not found error"}
	}
	profile := mdl.Profile{}
	err = dynamodbattribute.UnmarshalMap(getRes.Item, &profile)

	if err != nil {
		return "", err
	}
	return profile.MailBoxID, nil
}

func unmarshal(attribute map[string]events.DynamoDBAttributeValue, out interface{}) error {
	dbAttrMap := make(map[string]*dynamodb.AttributeValue)
	for k, v := range attribute {
		var dbAttr dynamodb.AttributeValue
		bytes, marshalErr := v.MarshalJSON()
		if marshalErr != nil {
			return nil
		}
		_ = json.Unmarshal(bytes, &dbAttr)
		dbAttrMap[k] = &dbAttr
	}
	return dynamodbattribute.UnmarshalMap(dbAttrMap, out)
}

type NotFoundError struct {
	Msg string
}

func (e NotFoundError) Error() string {
	return e.Msg
}
