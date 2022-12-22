package validators

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/fatmalabidi/protobuf/common"
	"github.com/fatmalabidi/protobuf/mailbox"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
)

// IsValidContext checks if a context has timeout or exceeded the default max timeout , returns bool
func IsValidContext(ctx context.Context, timeout time.Duration) bool {
	if ctx == nil {
		return false
	}
	if deadline, ok := ctx.Deadline(); !ok || deadline.Sub(time.Now()) > timeout*time.Second {
		return false
	}
	return true
}

// IsValidID validates is the string is an uuid.
func IsValidID(id string) bool {
	regex := regexp.MustCompile(`([0-9a-f]{8}-[0-9a-f]{4}-[0-5][0-9a-f]{3}-[089ab][0-9a-f]{3}-[0-9a-f]{12})`)
	result := regex.FindStringSubmatch(id)
	if result == nil {
		return false
	}
	return true
}

// IsValidListIDs validates if the list items are uuid.
func IsValidListIDs(ids []*common.MessageID) bool {
	for _, id := range ids {
		if !IsValidID(id.Value) {
			return false
		}
	}
	return true
}

// IsNilOrEmpty checks if any object of type defined by each case is nil or empty
func IsNilOrEmpty(input interface{}) bool {
	switch obj := input.(type) {
	case nil:
		return true
	case *common.UserID:
		return obj == nil || isStringEmpty(obj.Value)
	case *common.MailBoxID:
		return obj == nil || isStringEmpty(obj.Value)
	case *common.MessageID:
		return obj == nil || isStringEmpty(obj.Value)
	case *svc.GetMessageReq:
		return obj == nil || IsNilOrEmpty(obj.MessageID) || IsNilOrEmpty(obj.MailboxID)
	case *svc.SendMessageReq:
		return obj == nil || IsNilOrEmpty(obj.MailboxID) || IsNilOrEmpty(obj.MessageInfo)
	case *svc.DeleteMessagesReq:
		return obj == nil || IsNilOrEmpty(obj.MailboxID) || isListNilOrEmpty(obj.MessageIDs)
	case *svc.DeleteMessageReq:
		return obj == nil || IsNilOrEmpty(obj.MailboxID) || IsNilOrEmpty(obj.MessageID)
	case *svc.GetMessagesByParentReq:
		return obj == nil || IsNilOrEmpty(obj.MailboxID) || IsNilOrEmpty(obj.ParentMessageID)
	case *mailbox.MailMessageInfo:
		return obj == nil || IsNilOrEmpty(obj.SenderID) || IsNilOrEmpty(obj.RecipientID)
	default:
		return false
	}

}

// isStringEmpty checks if a string is empty, if it is, it returns true, otherwise it returns false
func isStringEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

//isListNilOrEmpty checks if all elements of a given list is nil or empty
func isListNilOrEmpty(list interface{}) bool {
	switch obj := list.(type) {
	case []*common.MessageID:
		for _, id := range obj {
			if id == nil || isStringEmpty(id.Value) {
				return true
			}
		}
	}
	return false
}
