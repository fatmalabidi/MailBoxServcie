package helpers_test

import (
	"context"
	td "github.com/fatmalabidi/MailBoxServcie/src/testdata/utils"
	. "github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"testing"
)

func TestPartitions(t *testing.T) {
	testcases := td.CreateTTPartition()
	for _, testcase := range testcases {
		t.Run(testcase.Name, func(t *testing.T) {
			var length int
			var idxRange IdxRange
			for idxRange = range Partition(testcase.CollectionLen, testcase.PartitionSize) {
				length++
			}
			if err := testcase.CheckResult(length, idxRange); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Errorf("expected a logger got nil")
	}
	testLogger := GetTestLogger()
	if testLogger == nil {
		t.Errorf("expected a test logger got nil")
	}
}

func TestAddTimeoutToCtx(t *testing.T) {
	t.Run("timeout < 1", func(t *testing.T) {
		ctx, _ := AddTimeoutToCtx(context.Background(), 1)
		if ctx.Err() != nil {
			t.Errorf("Expected Successful, recieved: %v\n", ctx.Err().Error())
		}
	})

	t.Run("timeout > 1", func(t *testing.T) {
		ctx, _ := AddTimeoutToCtx(context.Background(), 2)
		if ctx.Err() != nil {
			t.Errorf("Expected Successful, recieved: %v\n", ctx.Err().Error())
		}
	})
}
