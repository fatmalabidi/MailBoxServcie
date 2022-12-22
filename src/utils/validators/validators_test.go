package validators_test

import (
	"context"
	"github.com/bxcodec/faker/v3"
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/testdata/utils"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/validators"
	"github.com/fatmalabidi/protobuf/common"
	"github.com/google/uuid"
	"testing"
)

// Test_ValidContext tests if context is valid
func Test_ValidContext(t *testing.T) {
	conf, err := config.LoadConfig()
	if err != nil {
		t.Errorf("fail to load config for ValidContext test %v", err)
	}
	t.Run("valid context ", func(t *testing.T) {
		validContext, cancel := helpers.AddTimeoutToCtx(context.Background(), 1)
		defer cancel()
		if ok := validators.IsValidContext(validContext, conf.Server.Deadline); !ok {
			t.Error("expected valid context, got invalid")
		}

	})
	t.Run("invalid context ", func(t *testing.T) {
		invalidContext, cancelInvalid := helpers.AddTimeoutToCtx(context.Background(), 200000)
		defer cancelInvalid()
		if ok := validators.IsValidContext(invalidContext, conf.Server.Deadline); ok {
			t.Error("expected invalid context, got valid")
		}

	})

}

// Test_IsValidID tests IsValidID
func Test_IsValidID(t *testing.T) {
	t.Run("valid id", func(t *testing.T) {
		if ok := validators.IsValidID(uuid.New().String()); !ok {
			t.Error("expected valid ID, but was invalid")
		}
	})
	t.Run("invalid id", func(t *testing.T) {
		if ok := validators.IsValidID("invalid id"); ok {
			t.Error("expected invalid ID, but was valid")
		}
	})
}

func Test_IsValidListIDs(t *testing.T) {
	ids := []*common.MessageID{{Value: faker.UUIDHyphenated()}, {Value: faker.UUIDHyphenated()}}
	t.Run("valid IDs", func(t *testing.T) {
		if ok := validators.IsValidListIDs(ids); !ok {
			t.Error("expected valid IDs, got invalid")
		}
	})
	ids = []*common.MessageID{{Value: faker.UUIDHyphenated()}, {Value: "invalid"}}
	t.Run("invalid id", func(t *testing.T) {
		if ok := validators.IsValidID("invalid id"); ok {
			t.Error("expected invalid IDs, got valid")
		}
	})
}

func TestIsNilOrEmpty(t *testing.T) {
	for _, testCase := range utils.CreateTTIsNilOrEmpty() {
		t.Run(testCase.Name, func(t *testing.T) {
			if output := validators.IsNilOrEmpty(testCase.Object); output != testCase.IsNilOrEmpty {
				t.Errorf("expected %v ,got %v", testCase.IsNilOrEmpty, output)
			}
		})
	}
}
