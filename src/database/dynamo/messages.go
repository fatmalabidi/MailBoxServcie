package dynamo

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	mdl "github.com/fatmalabidi/MailBoxServcie/src/models"
	hlp "github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/sirupsen/logrus"
)

const BATCH_SIZE int = 20

// Repo is the handler for all db tables
type Repo struct {
	conf *config.Messages
	*dynamodb.DynamoDB
	log *logrus.Logger
}

// NewMessagesRepo creates a dynamo client and returns a Repo
func NewMessagesRepo(dbConf *config.Messages, logger *logrus.Logger) *Repo {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	dynamoClient := dynamodb.New(sess)
	dynamoHandler := Repo{
		conf:     dbConf,
		DynamoDB: dynamoClient,
		log:      logger,
	}
	return &dynamoHandler
}

// SendMessage adds the sent message to messages table
func (r *Repo) SendMessage(ctx context.Context, req *svc.SendMessageReq, ch chan<- mdl.SendMessageResponse) {
	r.newMessage(ctx, req, ch, true)
}

// GetMessage gets a specific message of a specific user
func (r *Repo) GetMessage(ctx context.Context, req *svc.GetMessageReq, ch chan<- mdl.GetMessageResponse) {
	// close the channel after sending the result
	defer close(ch)
	// check the request
	if req == nil {
		ch <- mdl.GetMessageResponse{
			Err: errors.New("request is null"),
		}
		return
	}
	// create get item input
	input := dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			mdl.MAILBOX_ID: {
				S: aws.String(req.MailboxID.Value),
			},
			mdl.MESSAGE_ID: {
				S: aws.String(req.MessageID.Value),
			},
		},
		TableName: aws.String(r.conf.TableName),
	}
	// get item from db
	result, err := r.GetItemWithContext(ctx, &input)
	// check error
	if err != nil {
		ch <- mdl.GetMessageResponse{
			Err: err,
		}
		return
	}
	// check result
	if result.Item == nil {
		ch <- mdl.GetMessageResponse{
			Err: errors.New("not found"),
		}
		return
	}
	// marshal the returned Item to a MessageModel
	message := mdl.MessageModel{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &message)
	if err != nil {
		ch <- mdl.GetMessageResponse{
			Err: err,
		}
		return
	}
	// return the message
	ch <- mdl.GetMessageResponse{
		Message: message,
	}
	return
}

// GetMessages gets all messages of a specific user
func (r *Repo) GetMessages(ctx context.Context, req *common.MailBoxID, ch chan<- mdl.GetMessageResponse) {
	r.getMessages(ctx, req, ch, SENT_AND_RECEIVED)
}

// GetSentMessages gets all the messages sent by a specific user
func (r *Repo) GetSentMessages(ctx context.Context, req *common.MailBoxID, ch chan<- mdl.GetMessageResponse) {
	r.getMessages(ctx, req, ch, SENT)
}

// GetReceivedMessages gets all the messages received by a specific user
func (r *Repo) GetReceivedMessages(ctx context.Context, req *common.MailBoxID, ch chan<- mdl.GetMessageResponse) {
	r.getMessages(ctx, req, ch, RECEIVED)
}

// GetMessagesByParent gets all the replies hierarchy of a specific message for a specific user
func (r *Repo) GetMessagesByParent(ctx context.Context, req *svc.GetMessagesByParentReq, ch chan<- mdl.GetMessageResponse) {
	// close the channel after sending the result
	defer close(ch)
	// check the request
	if req == nil {
		ch <- mdl.GetMessageResponse{
			Err: errors.New("request is null"),
		}
		return
	}
	// creating expression to filter results
	// specify the messages mailbox
	keyCon := expression.Key(mdl.MAILBOX_ID).Equal(expression.Value(req.MailboxID.Value))
	// filter by parentMessageID
	filter := expression.Name(mdl.PARENT_MESSAGE_ID).Equal(expression.Value(req.ParentMessageID.Value))
	// create an expression
	builder := expression.NewBuilder().WithKeyCondition(keyCon).WithFilter(filter)
	expr, err := builder.Build()
	if err != nil {
		ch <- mdl.GetMessageResponse{
			Err: err,
		}
		return
	}
	// get messages
	var items []mdl.MessageModel
	items, err = r.getMessagesByExpression(ctx, expr, nil, items)
	if err != nil {
		ch <- mdl.GetMessageResponse{Err: err}
		return
	}
	// send messages
	for _, message := range items {
		ch <- mdl.GetMessageResponse{
			Message: message,
		}
	}
	return
}

// MarkAsRead marks a message as read
func (r *Repo) MarkAsRead(ctx context.Context, req *svc.GetMessageReq, ch chan<- error) {
	// close the channel after sending the result
	defer close(ch)
	// check the request
	if req == nil {
		ch <- errors.New("request is null")
		return
	}
	// create update expression
	cond := expression.Key(mdl.MAILBOX_ID).Equal(expression.Value(req.MailboxID.Value)).
		And(expression.Key(mdl.MESSAGE_ID).Equal(expression.Value(req.MessageID.Value)))
	updateExpr := expression.Set(
		expression.Name(mdl.IS_READ),
		expression.Value(true))
	expr, err := expression.NewBuilder().WithUpdate(updateExpr).WithKeyCondition(cond).Build()
	if err != nil {
		ch <- err
		return
	}
	// create update item input
	input := dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			mdl.MAILBOX_ID: {
				S: aws.String(req.MailboxID.Value),
			},
			mdl.MESSAGE_ID: {
				S: aws.String(req.MessageID.Value),
			},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.KeyCondition(),
		TableName:                 aws.String(r.conf.TableName),
	}
	// update item in db
	_, err = r.UpdateItemWithContext(ctx, &input)
	// return result
	ch <- err
	return
}

// DeleteMessage deletes a message with/without its descendants from messages table
func (r *Repo) DeleteMessage(ctx context.Context, req *svc.DeleteMessageReq, ch chan<- error) {
	// close the channel after sending the result
	defer close(ch)
	// check WithDescendants option
	if req.WithDescendants {
		err := r.deleteMessageWithDescendants(ctx, req)
		ch <- err
		return
	}
	err := r.deleteMessageWithoutDescendants(ctx, req)
	ch <- err
	return
}

// deleteMessageWithoutDescendants deletes a message and its Descendants from messages table
func (r *Repo) deleteMessageWithDescendants(ctx context.Context, req *svc.DeleteMessageReq) error {
	// prepare a deleteMessages request
	deleteRequest := &svc.DeleteMessagesReq{
		MailboxID:  req.MailboxID,
		MessageIDs: []*common.MessageID{req.MessageID},
	}
	// get all descendants of the message
	// prepare channel
	getChan := make(chan mdl.GetMessageResponse)
	// prepare request
	getReq := &svc.GetMessagesByParentReq{
		MailboxID:       req.MailboxID,
		ParentMessageID: req.MessageID,
	}
	// submit GetMessageByParent request
	go r.GetMessagesByParent(ctx, getReq, getChan)
	for childMessage := range getChan {
		if childMessage.Err != nil {
			return childMessage.Err
		}
		deleteRequest.MessageIDs = append(deleteRequest.MessageIDs, &common.MessageID{Value: childMessage.Message.MessageID})
	}
	// create a channel and submit delete request
	delChan := make(chan error, 1)
	r.DeleteMessages(ctx, deleteRequest, delChan)
	// return result
	return <-delChan
}

// deleteMessageWithoutDescendants deletes one message from messages table
func (r *Repo) deleteMessageWithoutDescendants(ctx context.Context, req *svc.DeleteMessageReq) error {
	// create get item input
	input := dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			mdl.MAILBOX_ID: {
				S: aws.String(req.MailboxID.Value),
			},
			mdl.MESSAGE_ID: {
				S: aws.String(req.MessageID.Value),
			},
		},
		TableName: aws.String(r.conf.TableName),
	}
	// delete item from db
	_, err := r.DeleteItemWithContext(ctx, &input)
	return err
}

type awsBatch []*dynamodb.WriteRequest

// DeleteMessage deletes a list of messages from messages table
func (r *Repo) DeleteMessages(ctx context.Context, req *svc.DeleteMessagesReq, ch chan<- error) {
	// close the channel after sending the result
	defer close(ch)
	// check the request
	if req == nil {
		ch <- errors.New("request is null")
		return
	}
	// create a slice of dynamo DeleteItemInput
	batch := make(awsBatch, len(req.MessageIDs))
	// create the initial message structure
	for index, msgID := range req.MessageIDs {
		// create a message in dynamo type
		tobeDeleted := map[string]*dynamodb.AttributeValue{
			mdl.MAILBOX_ID: {
				S: aws.String(req.MailboxID.Value),
			},
			mdl.MESSAGE_ID: {
				S: aws.String(msgID.Value),
			},
		}
		// create delete input from tobeDeleted item
		writeRequest := dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{Key: tobeDeleted},
		}
		batch[index] = &writeRequest
	}
	// write deletion batch
	err := r.writeBatch(ctx, batch, 0)
	ch <- err
	return

}

// writeBatch writes the dynamo batch job after splitting the data into partitions of BATCH_SIZE
func (r *Repo) writeBatch(ctx context.Context, batch awsBatch, attempt int) error {
	for idxRange := range hlp.Partition(len(batch), BATCH_SIZE) {
		// create the batch array
		reqs := make([]*dynamodb.WriteRequest, BATCH_SIZE)
		reqs = append([]*dynamodb.WriteRequest(nil), batch[idxRange.Low:idxRange.High]...)
		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				r.conf.TableName: reqs,
			},
		}
		// write the deletion batch
		output, err := r.BatchWriteItemWithContext(ctx, input)
		// check error
		if err != nil {
			return err
		}
		// check if there is UnprocessedItems
		if output != nil && len(output.UnprocessedItems) != 0 {
			reWriteBatch := make(awsBatch, len(output.UnprocessedItems))
			reWriteBatch = append(reWriteBatch, output.UnprocessedItems[r.conf.TableName]...)
			// calculate waiting time (exponentially)
			if attempt != 0 {
				jitter := rand.Int63n(100)
				waitDuration := 1000*math.Pow(2, float64(attempt)) + float64(jitter)
				// wait before submitting again
				time.Sleep(time.Duration(waitDuration) * time.Millisecond)
			}
			// retry
			return r.writeBatch(ctx, reWriteBatch, attempt+1)
		}
	}
	return nil
}

// newMessage inserts a message to messages table, the outgoing parameter specifies if it is sent or received message
func (r *Repo) newMessage(ctx context.Context, req *svc.SendMessageReq, ch chan<- mdl.SendMessageResponse, outgoing bool) {
	// close the channel after sending the result
	defer close(ch)
	// check the request
	if req == nil {
		ch <- mdl.SendMessageResponse{
			Err: errors.New("request is null"),
		}
		return
	}
	// create a message model from the request info
	message := mdl.NewMessageFromRequest(req, outgoing)
	// marshal the model
	mailMessage, _ := dynamodbattribute.MarshalMap(message)
	// create an input item
	input := dynamodb.PutItemInput{
		Item:      mailMessage,
		TableName: aws.String(r.conf.TableName),
	}
	// insert item in db
	_, err := r.PutItemWithContext(ctx, &input)
	// check error
	if err != nil {
		ch <- mdl.SendMessageResponse{
			Err: err,
		}
		return
	}
	// send result
	ch <- mdl.SendMessageResponse{
		MessageID: message.MessageID,
		Err:       nil,
	}
	return
}

// ReceiveMessage adds the received message to messages table
func (r *Repo) ReceiveMessage(ctx context.Context, req *svc.SendMessageReq, ch chan<- mdl.SendMessageResponse) {
	r.newMessage(ctx, req, ch, false)
}

type MessageType int

const (
	SENT MessageType = iota
	RECEIVED
	SENT_AND_RECEIVED
)

// getMessages returns the list of messages in messages table of a specific mailbox,
// the outgoing parameter specifies if it is sent or received message
func (r *Repo) getMessages(ctx context.Context, req *common.MailBoxID, ch chan<- mdl.GetMessageResponse, outgoing MessageType) {
	// close the channel after sending the result
	defer close(ch)
	// check the request
	if req == nil {
		ch <- mdl.GetMessageResponse{
			Err: errors.New("request is null"),
		}
		return
	}
	// creating expression to filter results
	// specify the messages mailbox
	keyCon := expression.Key(mdl.MAILBOX_ID).Equal(expression.Value(req.Value))
	var builder expression.Builder
	// specify the message type (sent or received)
	switch outgoing {
	case SENT:
		filter := expression.Name(mdl.IS_OUT_GOING).Equal(expression.Value(true))
		builder = expression.NewBuilder().WithKeyCondition(keyCon).WithFilter(filter)
	case RECEIVED:
		filter := expression.Name(mdl.IS_OUT_GOING).Equal(expression.Value(false))
		builder = expression.NewBuilder().WithKeyCondition(keyCon).WithFilter(filter)
	default:
		builder = expression.NewBuilder().WithKeyCondition(keyCon)
	}
	expr, err := builder.Build()
	if err != nil {
		ch <- mdl.GetMessageResponse{
			Err: err,
		}
		return
	}
	// get messages
	var items []mdl.MessageModel
	items, err = r.getMessagesByExpression(ctx, expr, nil, items)
	if err != nil {
		ch <- mdl.GetMessageResponse{Err: err}
		return
	}
	// send messages
	for _, message := range items {
		ch <- mdl.GetMessageResponse{
			Message: message,
		}
	}
	return
}

// getMessagesByExpression returns messages list in messages table using the expression passed in parameter
func (r *Repo) getMessagesByExpression(ctx context.Context, expr expression.Expression, lastItem map[string]*dynamodb.AttributeValue, items []mdl.MessageModel) ([]mdl.MessageModel, error) {
	// creating a QueryInput
	queryInput := &dynamodb.QueryInput{
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(r.conf.TableName),
		ExclusiveStartKey:         lastItem,
	}
	// execute query
	output, err := r.QueryWithContext(ctx, queryInput)
	// check error
	if err != nil {
		return items, err
	}
	// unmarshal results
	for _, message := range output.Items {
		messageModel := mdl.MessageModel{}
		err = dynamodbattribute.UnmarshalMap(message, &messageModel)
		if err != nil {
			return items, err
		}
		items = append(items, messageModel)
	}
	// check if there is other unloaded results
	if output.LastEvaluatedKey != nil {
		return r.getMessagesByExpression(ctx, expr, output.LastEvaluatedKey, items)
	}
	return items, nil
}
